"""Send all five booking requests and verify their traces in Jaeger."""

import argparse
import datetime
import json
import time
import urllib.error
import urllib.parse
import urllib.request
import uuid

BOOKINGS = "http://127.0.0.1:8081/api/v1"
JAEGER = "http://127.0.0.1:16686"
MAILPIT = "http://127.0.0.1:8025"


def request(url, method="GET", body=None, headers=None, status=200):
    headers = dict(headers or {})
    data = None
    if body is not None:
        headers["Content-Type"] = "application/json"
        data = json.dumps(body).encode()
    req = urllib.request.Request(url, data=data, headers=headers, method=method)
    with urllib.request.urlopen(req, timeout=10) as response:
        assert response.status == status, (url, response.status, status)
        payload = response.read()
        return json.loads(payload) if payload else None


def booking_request(path, **kwargs):
    trace_id = uuid.uuid4().hex
    caller_id = uuid.uuid4().hex[:16]
    headers = {"traceparent": f"00-{trace_id}-{caller_id}-01", "tracestate": "test=booking"}
    return request(BOOKINGS + path, headers=headers, **kwargs), trace_id, caller_id


def tags(span):
    return {tag["key"]: tag["value"] for tag in span["tags"]}


def verify(trace, operation, trace_id, caller_id, http_status):
    spans = trace["spans"]
    processes = trace["processes"]

    def service(span):
        return processes[span["processID"]]["serviceName"]

    def find(name, svc, kind=None):
        matches = [
            span for span in spans
            if span["operationName"] == name
            and service(span) == svc
            and (kind is None or tags(span).get("span.kind") == kind)
        ]
        assert len(matches) == 1, (name, svc, kind, len(matches))
        return matches[0]

    def parent(child, parent_id):
        assert any(
            ref["refType"] == "CHILD_OF"
            and ref["traceID"] == trace_id
            and ref["spanID"] == parent_id
            for ref in child["references"]
        ), (child["operationName"], child["references"], parent_id)

    def queries(svc, parent_span):
        matches = [
            span for span in spans
            if service(span) == svc
            and span["operationName"] in {"sql.conn.query", "sql.conn.exec"}
        ]
        assert matches, f"No PostgreSQL spans from {svc}"
        for span in matches:
            parent(span, parent_span["spanID"])
            assert tags(span).get("db.system.name") == "postgresql"
            assert "db.query.text" not in tags(span) and "db.statement" not in tags(span)

    assert all(span["traceID"] == trace_id for span in spans)
    http = find(operation, "bookingsvc", "server")
    assert tags(http).get("http.response.status_code") == http_status
    parent(http, caller_id)
    business = find("BookingService." + operation, "bookingsvc")
    parent(business, http["spanID"])
    queries("bookingsvc", business)

    rpc = {"GetAvailableRooms": "GetAll", "CreateBooking": "Get"}.get(operation)
    if rpc:
        rpc_name = "rooms.Rooms/" + rpc
        client = find(rpc_name, "bookingsvc", "client")
        server = find(rpc_name, "hotelsvc", "server")
        parent(client, business["spanID"])
        parent(server, client["spanID"])
        queries("hotelsvc", server)

    event = {"CreateBooking": "booking.created", "CancelBooking": "booking.cancelled"}.get(operation)
    if event:
        producer = find("publish bookings.events", "bookingsvc", "producer")
        consumer = find("process bookings.events", "notificationsvc", "consumer")
        notification = find(event + " notify", "notificationsvc")
        smtp = find("SMTP send", "notificationsvc", "client")
        parent(producer, business["spanID"])
        parent(consumer, producer["spanID"])
        parent(notification, consumer["spanID"])
        parent(smtp, notification["spanID"])
        assert tags(consumer).get("event.name") == event
        assert tags(consumer).get("messaging.process.result") == "success"
        assert tags(smtp).get("smtp.attempt") == 1 and not tags(smtp).get("error")


def wait_for_trace(operation, trace_id, caller_id, status):
    last_error = None
    for _ in range(45):
        try:
            traces = request(f"{JAEGER}/api/traces/{trace_id}")["data"]
            assert traces, "Jaeger returned no trace"
            verify(traces[0], operation, trace_id, caller_id, status)
            print(f"{operation} HTTP {status} {trace_id} ({len(traces[0]['spans'])} spans)", flush=True)
            return
        except (urllib.error.HTTPError, AssertionError) as error:
            last_error = error
            time.sleep(2)
    raise RuntimeError(f"{operation} trace {trace_id} failed verification") from last_error


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("hotel_id", type=int, help="Hotel with an available test room")
    args = parser.parse_args()
    request(MAILPIT + "/api/v1/messages")
    date_from = datetime.date.today() + datetime.timedelta(days=30)
    date_to = date_from + datetime.timedelta(days=2)
    dates = {"date_from": date_from.isoformat(), "date_to": date_to.isoformat()}

    available, trace, caller = booking_request(
        f"/hotels/{args.hotel_id}/available-rooms?" + urllib.parse.urlencode(dates)
    )
    checks = [("GetAvailableRooms", trace, caller, 200)]
    assert available["rooms"], "No available room for the test dates"
    created, trace, caller = booking_request(
        "/bookings", method="POST", status=201,
        body={"user_id": 1, "hotel_id": args.hotel_id, "room_id": available["rooms"][0]["id"], **dates},
    )
    booking_id = created["id"]
    checks.append(("CreateBooking", trace, caller, 201))
    listed, trace, caller = booking_request(f"/bookings?hotel_id={args.hotel_id}")
    assert any(booking["id"] == booking_id for booking in listed["bookings"])
    checks.append(("GetBookings", trace, caller, 200))
    fetched, trace, caller = booking_request(f"/bookings/{booking_id}")
    assert fetched["id"] == booking_id
    checks.append(("GetBookingByID", trace, caller, 200))
    _, trace, caller = booking_request(f"/bookings/{booking_id}", method="DELETE", status=204)
    checks.append(("CancelBooking", trace, caller, 204))

    for check in checks:
        wait_for_trace(*check)

    messages = request(MAILPIT + "/api/v1/messages")["messages"]
    subjects = {message["Subject"] for message in messages}
    for suffix in ("подтверждено", "отменено"):
        assert f"Бронирование №{booking_id} {suffix}" in subjects
    print(f"Mailpit received both notifications for booking {booking_id}")


if __name__ == "__main__":
    main()

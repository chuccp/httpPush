"""
WebSocket end-to-end test: HTTP API -> WS receive
Usage: python script/test_ws.py
"""
import threading
import time
import requests
import websocket

HOST = "http://127.0.0.1:8086"
WS_URL = "ws://127.0.0.1:8086/ws?id=user_a"

received = []

def ws_listen():
    ws = websocket.create_connection(WS_URL)
    print(f"[WS] connected")
    try:
        ws.settimeout(1)
        try:
            ws.recv()
        except:
            pass

        ws.settimeout(5)
        while True:
            msg = ws.recv()
            received.append(msg)
            print(f"[WS] received: {msg}")
    except websocket.WebSocketTimeoutException:
        print("[WS] timeout, closing")
    except Exception as e:
        print(f"[WS] closed: {e}")
    finally:
        ws.close()

def http_send(to="user_a", msg="hello-from-http"):
    url = f"{HOST}/sendmsg?username={to}&msg={msg}"
    print(f"[HTTP] GET {url}")
    resp = requests.get(url)
    print(f"[HTTP] response: {resp.status_code} {resp.text}")

if __name__ == "__main__":
    t = threading.Thread(target=ws_listen, daemon=True)
    t.start()
    time.sleep(1)

    http_send("user_a", "hello from http api")

    time.sleep(3)

    if received:
        print(f"\nPASS: WS received {len(received)} messages")
    else:
        print("\nFAIL: WS received nothing")

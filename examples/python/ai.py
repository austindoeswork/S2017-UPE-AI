# pip install ws4py
from ws4py.client.threadedclient import WebSocketClient

url = "ws://npcompete.io/wsplay"
key = "IngesFastFaultyBigtoe"

class WSBot(WebSocketClient):
    def opened(self):
        print "sending key"
        self.send(key)

    def closed(self, code, reason=None):
        print "Closed down", code, reason

    def received_message(self, m):
        print "recvd:", m

if __name__ == '__main__':
    try:
        ws = WSBot(url, protocols=['http-only'])
        ws.connect()
        ws.run_forever()
    except KeyboardInterrupt:
        ws.close()

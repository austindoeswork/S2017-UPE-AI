var ws;
window.addEventListener("load", function(evt) {
    var output = document.getElementById("output");
    var input = document.getElementById("input");
    document.getElementById("open").onclick = function(evt) {
		var wspath = document.getElementById("wspathinput").value;
		var gname = document.getElementById("gamename").value;
        if (ws) {
            return false;
        }

		console.log(wspath);
        ws = new WebSocket(wspath + "?game=" + gname);
        ws.onopen = function(evt) {
            console.log("WS CONNECTION OPENED");
        }
        ws.onclose = function(evt) {
            console.log("WS CONNECTION CLOSED");
            ws = null;
        }
        ws.onmessage = function(evt) {
			//console.log(evt);
			renderGrid(evt.data);
        }
        ws.onerror = function(evt) {
            console.log("ERR: " + evt.data);
        }
        return false;
    };
    document.getElementById("send").onclick = function(evt) {
		return send(evt.data)
    };
    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
		console.log("CLOSING WS");
        ws.close();
        return false;
    };

});

function send(input) {
	console.log("sending");
        if (!ws) {
	console.log("failed");
            return false;
        }
        ws.send(input);
	console.log("success");
        return false;
}

var ws;
window.addEventListener("load", function(evt) {
    document.getElementById("join").onclick = wsjoin;
    document.getElementById("watch").onclick = wswatch;
    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
		console.log("CLOSING WS");
        ws.close();
        return false;
    };
});

function wsjoin() {
	var wspath = document.getElementById("wspathinput").value;
	var wsroute = "/wsjoin"
	var gname = document.getElementById("gamename").value;
	wsopen(wspath, wsroute, gname);
}
function wswatch() {
	var wspath = document.getElementById("wspathinput").value;
	var wsroute = "/wswatch"
	var gname = document.getElementById("gamename").value;
	wsopen(wspath, wsroute, gname);
}

function wsopen(wspath, wsroute, gname) {
	if (ws) {
		return false;
	}
	ws = new WebSocket(wspath + wsroute + "?game=" + gname);
	ws.onopen = function(evt) {
		setstatus("WS CONNECTION OPENED");
		console.log("WS CONNECTION OPENED");
	}
	ws.onclose = function(evt) {
		setstatus("WS CONNECTION CLOSED");
		console.log("WS CONNECTION CLOSED");
		console.log(evt);
		ws = null;
	}
	ws.onmessage = function(evt) {
		renderGrid(evt.data);
	}
	ws.onerror = function(evt) {
		console.log(evt);
		setstatus(evt.data);
	}
	return false;
}
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
function setstatus(statusstring) {
	s = document.getElementById("status");
	s.innerHTML = "CONNECTION OPENED";
}

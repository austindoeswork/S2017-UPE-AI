var ws;
window.addEventListener("load", function(evt) {
    console.log("LOADING");
    document.getElementById("join").onclick = wsjoin;
    document.getElementById("play").onclick = wsplay;
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

function wsplay() {
    console.log("play clicked");
	var wspath = document.getElementById("wspathinput").value;
	var wsroute = "/wsplay";
	// var gname = document.getElementById("gamename").value;
	var devkey = document.getElementById("devkey").value;
	wsopen(wspath, wsroute, "", devkey);
}
function wsjoin() {
	console.log("join clicked");
	var wspath = document.getElementById("wspathinput").value;
	var wsroute = "/wsjoin";
	var gname = document.getElementById("gamename").value;
	var devkey = document.getElementById("devkey").value;
	wsopen(wspath, wsroute, gname, devkey);
}
function wswatch() {
	var wspath = document.getElementById("wspathinput").value;
	var wsroute = "/wswatch"
	var gname = document.getElementById("gamename").value;
	wsopen(wspath, wsroute, gname, "");
}

function wsopen(wspath, wsroute, gname, devkey) {
	if (ws) {
		return false;
	}
	if (gname != "") {
		ws = new WebSocket(wspath + wsroute + "?game=" + gname);
	} else {
		ws = new WebSocket(wspath + wsroute);
	}

	ws.onopen = function(evt) {
		if (devkey != "") {
			send(devkey);
		}
		setstatus("WS send devkey");
		console.log("WS send devkey");
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
	s.innerHTML = statusstring;
}

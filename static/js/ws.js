var ws;
window.addEventListener("load", function(evt) {
    document.getElementById("join").onclick = wsjoin;
    document.getElementById("play").onclick = wsplay;
    document.getElementById("watch").onclick = wswatch;
    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
	setstatus("NOT CONNECTED", "label label-danger");
	ws.close();
        return false;
    };
});

function wsplay() {
    var wspath = document.getElementById("wspathinput").value;
    var wsroute = "/wsplay";
    // var gname = document.getElementById("gamename").value;
    var devkey = document.getElementById("devkey").value;
    wsopen(wspath, wsroute, "", devkey);
}
function wsjoin() {
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
	    setstatus("Sent devkey to server", "label label-info");
	}
	else {
	    setStatus("Please enter your devkey", "label label-warning");
	}
    }
    ws.onclose = function(evt) {
	setstatus("NOT CONNECTED", "label label-danger");
	ws = null;
    }
    ws.onmessage = function(evt) {
	setstatus("GAME CONNECTED", "label label-success");
	renderGrid(evt.data);
    }
    ws.onerror = function(evt) {
	console.log(evt);
	setstatus(evt.data);
    }
    return false;
}
function send(input) {
    if (!ws) {
	return false;
    }
    ws.send(input);
    return false;
}
function setstatus(statusstring, className) {
    s = document.getElementById("status");
    s.className = className;
    s.innerHTML = statusstring;
}

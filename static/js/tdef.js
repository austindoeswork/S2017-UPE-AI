var mycanvas;
xbuffer = 10
canvasw = 30 + 2 * xbuffer;
canvash = 20;

pwidth = 3;
plength = 5;

p1x = -1;
p1hp = -1;
p2x = -1;
p2hp = -1;

units = [];

function setup(){
    mycanvas = createCanvas(canvasw, canvash);
    mycanvas.parent('canvasHolder')
    background(100);
    noLoop();
}

function buyTower(location) {
    var radioButtons = document.getElementsByName('towerEnum');
    var enumVal;
    for(var i = 0; i < radioButtons.length; i++){
	if(radioButtons[i].checked){
            enumVal = radioButtons[i].value;
	}
    }
    // console.log('b' + enumVal + ' ' + location)
    send('b' + enumVal + ' ' + location)
}

function draw(){
    background(100);

    // units
    for (i = 0; i < units.length; i++) {
	var c;
	if (units[i].hp/units[i].maxhp > .5) {
	    c = color('green');
	}
	else if (units[i].hp/units[i].maxhp > .3) {
	    c = color('yellow');
	}
	else {
	    c = color('red');
	}
	fill(c);
	if (units[i].enum == -2) { // temporary workaround until we add ids or something
	    rect(xbuffer + units[i].x-5, canvash - units[i].y+40, 10, 200);
	}
	else if (units[i].enum == -1) {
	    rect(xbuffer + units[i].x-5, canvash - units[i].y+40, 10, 100);
	}
	else if (units[i].enum == 0) {
	    rect(xbuffer + units[i].x-5, canvash - units[i].y+40, 10, 40);
	}
	else if (units[i].enum >= 10) {
	    rect(xbuffer + units[i].x - 50, canvash - units[i].y+40, 50, 50);
	}
    }
}

function renderGrid(data) {
    d = JSON.parse(data);
    if (canvasw != d.w+2*xbuffer || canvash != d.h) { // ?
	canvasw = d.w+2*xbuffer
	canvash = d.h
	mycanvas.size(canvasw, canvash)
    }
    p1x = d.p1.mainTower.x;
    p2x = d.p2.mainTower.x;
    units = d.p1.units.concat(d.p2.units);
    units.push(d.p1.mainTower);
    units.push(d.p2.mainTower);
    document.getElementById("p1hp").innerHTML = d.p1.mainTower.hp
    document.getElementById("p2hp").innerHTML = d.p2.mainTower.hp
    document.getElementById("p1coins").innerHTML = d.p1.coins
    document.getElementById("p2coins").innerHTML = d.p2.coins
    redraw();
}

new p5();

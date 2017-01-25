var mycanvas;

innerWidth = 800;
innerHeight = 300;

canvasw = 1600;
canvash = 600;

pwidth = 3;
plength = 5;

p1x = -1;
p1hp = -1;
p2x = -1;
p2hp = -1;

units = [];

function scaleX(oldX) {
    return oldX * innerWidth / canvasw;
}

function scaleY(oldY) {
    return oldY * innerHeight / canvash;
}

function setup(){
    mycanvas = createCanvas(innerWidth, innerHeight);
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
	    rect(scaleX(units[i].x-5), scaleY(canvash - units[i].y+40), 10, 200);
	}
	else if (units[i].enum == -1) {
	    rect(scaleX(units[i].x-5), scaleY(canvash - units[i].y+40), 10, 100);
	}
	else if (units[i].enum == 0) {
	    rect(scaleX(units[i].x-5), scaleY(canvash - units[i].y+40), 10, 40);
	}
	else if (units[i].enum >= 50) {
	    rect(scaleX(units[i].x-5), scaleY(canvash - units[i].y+40), 20, 20);
	}
    }
}

function renderGrid(data) {
    console.log(data);
    d = JSON.parse(data);
    if (canvasw != d.w || canvash != d.h) { // ?
	canvasw = d.w;
	canvash = d.h;
	mycanvas.size(canvasw, canvash);
    }
    p1x = d.p1.mainTower.x;
    p2x = d.p2.mainTower.x;
    units = d.p1.troops.concat(d.p2.troops);
    for (i = 0; i < d.p1.towers.length; i++) {
	if (d.p1.towers[i] != 'nil') {
	    units.push(d.p1.towers[i]);
	}
    }
    for (i = 0; i < d.p2.towers.length; i++) {
	if (d.p2.towers[i] != 'nil') {
	    units.push(d.p2.towers[i]);
	}
    }
    units.push(d.p1.mainTower);
    units.push(d.p2.mainTower);
    document.getElementById("p1hp").innerHTML = d.p1.mainTower.hp
    document.getElementById("p2hp").innerHTML = d.p2.mainTower.hp
    document.getElementById("p1bits").innerHTML = d.p1.bits
    document.getElementById("p2bits").innerHTML = d.p2.bits
    document.getElementById("p1income").innerHTML = d.p1.income
    document.getElementById("p2income").innerHTML = d.p2.income
    redraw();
}

new p5();

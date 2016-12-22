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
      if (units[i].maxhp == 1000) { // temporary workaround until we add ids or something
	  rect(xbuffer + units[i].x, 10, 2, 10);
      }
      else {
	  rect(xbuffer + units[i].x, 15, 2, 5);
      }
  }
}

function renderGrid(data) {
    // console.log(data);
    d = JSON.parse(data);
    // console.log(d);
    if (canvasw != d.w+2*xbuffer || canvash != d.h) { // ?
	  canvasw = d.w+2*xbuffer
	  canvash = d.h
	  mycanvas.size(canvasw, canvash)
    }
    p1x = d.p1.x;
    p2x = d.p2.x;
    units = d.units;
    // console.log(units);
    document.getElementById("p1hp").innerHTML = d.p1.hp
    document.getElementById("p2hp").innerHTML = d.p2.hp
    redraw();
}

/* function renderGrid(data) {
  d = JSON.parse(data);
  console.log(d);
  if (canvasw != d.w+2*xbuffer || canvash != d.h) {
	  canvasw = d.w+2*xbuffer
	  canvash = d.h
	  mycanvas.size(canvasw, canvash)
  }
	
  ballx = d.bx;
  bally = d.by;

  p1x = d.p1x;
  p1y = d.p1y;
  p1s = d.p1s;
  p2x = d.p2x;
  p2y = d.p2y;
  p2s = d.p2s;
  plength = d.l;

  document.getElementById("p1score").innerHTML = p1s
  document.getElementById("p2score").innerHTML = p2s

  redraw();

} */

// 38 up
// 37 left
// 40 down
// 39 right
function keyPressed() {
  console.log("p5 key: " + keyCode);
  if (keyCode == 38) {
	  send("up");
  }
  else if (keyCode == 40) {
	  send("down");
  }
}

new p5();

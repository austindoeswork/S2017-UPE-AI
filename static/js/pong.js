var mycanvas;
xbuffer = 10
canvasw = 30 + 2 * xbuffer;
canvash = 20;


ballx = canvasw/2 - xbuffer;
bally = canvash/2;

pwidth = 3;
plength = 5;

p1x = -100;
p1y = -100;

p2x = -100;
p2y = -100;

function setup(){
  mycanvas = createCanvas(canvasw, canvash);
  mycanvas.parent('canvasHolder')
  background(100);
  noLoop();
}

function draw(){
  background(100);

  // ball
  rect(xbuffer + ballx-1, bally-1, 3, 3);

  // p1
  rect(xbuffer + p1x - pwidth + 1, p1y, pwidth, plength);
  // p2
  rect(xbuffer + p2x, p2y, pwidth, plength);
}


function renderGrid(data) {
  d = JSON.parse(data);
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

}

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

// TODO turn this into a generalized gameTV creator that can make smaller windows (i.e. for the main page)

// scaleFactor scales the gameWindow to browser screen
var scaleFactor = 1;
var GAME_WIDTH = 1600;
var GAME_HEIGHT = 600;

// pixiJS lets us use whichever from webgl/canvas is enabled
var type = "WebGL"
if(!PIXI.utils.isWebGLSupported()){
    type = "canvas"
}

PIXI.utils.sayHello(type) // not really necessary, but could be nice for people to know

//Aliases
var Container = PIXI.Container;
var autoDetectRenderer = PIXI.autoDetectRenderer;
var loader = PIXI.loader;
var resources = PIXI.loader.resources;
var Sprite = PIXI.Sprite;
var TextureCache = PIXI.utils.TextureCache;
var rendererOptions = {
    antialiasing: false,
    transparent: false,
    resolution: window.devicePixelRatio,
    autoResize: true,
};

var renderer = autoDetectRenderer(GAME_WIDTH, GAME_HEIGHT, rendererOptions);
renderer.backgroundColor = 0xCCEBF1; // baby blue
document.getElementById('gameTV').appendChild(renderer.view);
var stage = new Container(); // TODO: make into particlecontainer?

function resize() { // autoresizes gameTV depending on size of window (which determines size of main container)
    // Determine which screen dimension is most constrained
    // note that we use main container for height because gameTV is not a fixed size
    scaleFactor = Math.min(document.getElementById('gameTV').offsetWidth/GAME_WIDTH,
			   document.getElementById('main').offsetHeight/GAME_HEIGHT);
    
    // Scale the view appropriately to fill that dimension
    stage.scale.x = stage.scale.y = scaleFactor;

    // Update the renderer dimensions
    renderer.resize(Math.ceil(GAME_WIDTH * scaleFactor),
                    Math.ceil(GAME_HEIGHT * scaleFactor));
    renderer.render(stage);
}

// load images as textures, and once they're loaded, run setup
loader
    .add([
	"static/img/background.png",
	"static/img/Mob_2_Red.png",
	"static/img/Mob_2_Blue.png"
    ])
    .load(setup);

var Mob2Blues = []; // prerender 100 of these
var Mob2Reds = []; // prerender 100 of these as well

// runs as soon as loader is done loading imgs
function setup() {
    stage.addChild(new Sprite(
	TextureCache["static/img/background.png"]
    ));
    for (var i = 0; i < 50; i++) {
	var mob = new Sprite(
	    TextureCache["static/img/Mob_2_Blue.png"]
	);
	mob.x = -100;
	mob.y = -100;
	Mob2Blues.push(mob);
	stage.addChild(mob);
	mob = new Sprite(
	    TextureCache["static/img/Mob_2_Red.png"]
	);
	mob.x = -100;
	mob.y = -100;
	mob.scale.x = -1; // flip horizontally
	Mob2Reds.push(mob);
	stage.addChild(mob);
    }
    // stage.addChild(particleContainer);
    resize();
    renderer.render(stage);
}

// resize when user changes page size
window.addEventListener("resize", resize);

// called by frontend troop buttons
function buyTroop(location) {
    var radioButtons = document.getElementsByName('troopEnum');
    var enumVal;
    for(var i = 0; i < radioButtons.length; i++){
	if(radioButtons[i].checked){
            enumVal = radioButtons[i].value;
	}
    }
    send('b' + enumVal + ' ' + location)
}

// called by frontend tower buttons
function buyTower(location) {
    var radioButtons = document.getElementsByName('towerEnum');
    var enumVal;
    for(var i = 0; i < radioButtons.length; i++){
	if(radioButtons[i].checked){
            enumVal = radioButtons[i].value;
	}
    }
    send('b' + enumVal + ' ' + location)
}

var timestamp = Date.now();

function renderGrid(data) {
    frames++;
    d = JSON.parse(data);
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
    document.getElementById("p1hp").innerHTML = d.p1.mainTower.hp;
    document.getElementById("p2hp").innerHTML = d.p2.mainTower.hp;
    document.getElementById("p1bits").innerHTML = d.p1.bits;
    document.getElementById("p2bits").innerHTML = d.p2.bits;
    document.getElementById("p1income").innerHTML = d.p1.income;
    document.getElementById("p2income").innerHTML = d.p2.income;
    draw(units);
    renderer.render(stage);
    document.getElementById("fps").innerHTML = 1000/(Date.now() - timestamp);
    timestamp = Date.now();
}

function draw(units){
    var mobBlueIterator = 0;
    var mobRedIterator = 0;

    for (var test = 0; test < 50; test++) { // reset all of the sprites
	Mob2Blues[test].x = -100;
	Mob2Blues[test].y = -100;
	Mob2Reds[test].x = -100;
	Mob2Reds[test].y = -100;
    }
    
    // units
    for (var i = 0; i < units.length; i++) {
	var thisUnit;
	if (units[i].owner == 1) {
	    thisUnit = Mob2Blues[mobBlueIterator];
	    mobBlueIterator++;
	}
	else {
	    thisUnit = Mob2Reds[mobRedIterator];
	    mobRedIterator++;
	}
	
	if (units[i].enum == -1) {
	    thisUnit.x = units[i].x;
	    thisUnit.y = GAME_HEIGHT - units[i].y;
	}
	else if (units[i].enum >= 0 && units[i].enum < 50) {
	    thisUnit.x = units[i].x;
	    thisUnit.y = GAME_HEIGHT - units[i].y;
	}
	else if (units[i].enum >= 50) {
	    thisUnit.x = units[i].x;
	    thisUnit.y = GAME_HEIGHT - units[i].y;
	}
    }
}

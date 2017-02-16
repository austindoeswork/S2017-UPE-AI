// TODO turn this into a generalized gameTV creator that can make smaller windows (i.e. for the main page)

function buttonPress(button) {
    buyTower(button.id);
}

// populate game controller
function populateGameController() {
    for (var i = 0; i < 66; i++) {
	if (i > 0 && i%11 == 0) {
	    document.getElementById('towerController').append(document.createElement("br"));
	}
	var newButton = document.createElement("button");
	newButton.className = "btn btn-default";
	var towerenum = i;
	if (towerenum < 10) {
	    towerenum = '0' + i;
	}
	newButton.id = towerenum;
	newButton.onclick = function() {
	    buttonPress(this);
	}
	newButton.innerHTML = towerenum;
	document.getElementById('towerController').appendChild(newButton);
    }
}

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

// TODO: try to figure out how to load these dynamically, just want to get the prototype out
// load images as textures, and once they're loaded, run setup
var readyToDisplay = false;
loader
    .add("background", "static/img/background.png")
    .add("blueLaneColors", "static/img/Lane_colors_Blue.png")
    .add("redLaneColors", "static/img/Lane_colors_Red.png")
    .add("blueBuildingBar", "static/img/Building_barcolor_Blue.png")
    .add("redBuildingBar", "static/img/Building_barcolor_Red.png")
    .add("steps", "static/img/Building_steps1.png")
    .add("-1-blue", "static/img/Tower1_Blue.png")
    .add("-1-red", "static/img/Tower1_Red.png")
    .add("0-blue", "static/img/nut1.png")
    .add("0-red", "static/img/nut2.png")
    .add("1-blue", "static/img/bolts1.png")
    .add("1-red", "static/img/bolts2.png")
    .add("2-blue", "static/img/GreaseMonkey1.png")
    .add("2-red", "static/img/GreaseMonkey2.png")
    .add("3-blue", "static/img/walker1.png")
    .add("3-red", "static/img/walker2.png")
    .add("4-blue", "static/img/aimbot1.png")
    .add("4-red", "static/img/aimbot2.png")
    .add("5-blue", "static/img/hardrive1.png")
    .add("5-red", "static/img/hardrive2.png")
    .add("6-blue", "static/img/scrapheap1.png")
    .add("6-red", "static/img/scrapheap2.png")
    .add("7-blue", "static/img/gasguzzler1.png")
    .add("7-red", "static/img/gasguzzler2.png")
    .add("8-blue", "static/img/terminator1.png")
    .add("8-red", "static/img/terminator2.png")
    .add("9-blue", "static/img/blackhat1.png")
    .add("9-red", "static/img/blackhat2.png")
    .add("10-blue", "static/img/malware1.png")
    .add("10-red", "static/img/malware2.png")
    .add("11-blue", "static/img/ghandi1.png")
    .add("11-red", "static/img/ghandi2.png")
    .add("50-blue", "static/img/peashooter1.png")
    .add("50-red", "static/img/peashooter2.png")
    .add("51-blue", "static/img/firewall1.png")
    .add("51-red", "static/img/firewall2.png")
    .add("52-blue", "static/img/guardian1.png")
    .add("52-red", "static/img/guardian2.png")
    .add("53-blue", "static/img/bank.png")
    .add("53-red", "static/img/bank.png")
    .add("54-blue", "static/img/junkyard1.png")
    .add("54-red", "static/img/junkyard2.png")
    .add("55-blue", "static/img/startup1.png")
    .add("55-red", "static/img/startup2.png")
    .add("56-blue", "static/img/corporation1.png")
    .add("56-red", "static/img/corporation2.png")
    .add("57-blue", "static/img/warpdrive1.png")
    .add("57-red", "static/img/warpdrive2.png")
    .add("58-blue", "static/img/jammingstation1.png")
    .add("58-red", "static/img/jammingstation2.png")
    .add("59-blue", "static/img/Hotspot1.png")
    .add("59-red", "static/img/Hotspot2.png")
    .load(setup);


// we render all the mobs we need on a need-basis
var prerenderedMobs = {};
for (var i = -1; i <= 11; i++) {
    prerenderedMobs['' + i + '-blue'] = [];
    prerenderedMobs['' + i + '-red'] = [];
}
for (var i = 50; i <= 59; i++) {
    prerenderedMobs['' + i + '-blue'] = [];
    prerenderedMobs['' + i + '-red'] = [];
}
var prerenderedMobIterators = {};
for (var i = -1; i <= 11; i++) {
    prerenderedMobIterators['' + i + '-blue'] = 0;
    prerenderedMobIterators['' + i + '-red'] = 0;
}
for (var i = 50; i <= 59; i++) {
    prerenderedMobIterators['' + i + '-blue'] = 0;
    prerenderedMobIterators['' + i + '-red'] = 0;
}

// runs as soon as loader is done loading imgs
function setup() {
    // create the background, this will never get changed
    stage.addChild(new Sprite(
	TextureCache["background"]
    ));
    var blueLaneColors = new Sprite(
	TextureCache["blueLaneColors"]
    );
    blueLaneColors.scale.x = -1;
    blueLaneColors.x = blueLaneColors.width + 15;
    blueLaneColors.y = GAME_HEIGHT - blueLaneColors.height - 50;
    stage.addChild(blueLaneColors);
    var redLaneColors = new Sprite(
	TextureCache["redLaneColors"]
    );
    redLaneColors.x = GAME_WIDTH - redLaneColors.width - 15;
    redLaneColors.y = GAME_HEIGHT - redLaneColors.height - 50;
    stage.addChild(redLaneColors);
    var blueBuildingBar = new Sprite(
	TextureCache["blueBuildingBar"]
    );
    blueBuildingBar.x = blueBuildingBar.width;
    blueBuildingBar.scale.x = -1;
    blueBuildingBar.y = 15;
    stage.addChild(blueBuildingBar);
    var redBuildingBar = new Sprite(
	TextureCache["redBuildingBar"]
    );
    redBuildingBar.x = GAME_WIDTH - redBuildingBar.width;
    redBuildingBar.y = 15;
    stage.addChild(redBuildingBar);
    var leftSteps = new Sprite(
	TextureCache["steps"]
    );
    leftSteps.scale.x = -1;
    leftSteps.x = leftSteps.width - 10; leftSteps.y = 0;
    stage.addChild(leftSteps);
    var rightSteps = new Sprite(
	TextureCache["steps"]
    );
    rightSteps.x = GAME_WIDTH - rightSteps.width + 10;
    rightSteps.y = 0;
    stage.addChild(rightSteps);
    resize();
    renderer.render(stage);
    readyToDisplay = true;
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
    units = d.p1.troops.concat(d.p2.troops); // TODO: make it so the top towers are drawn first, then the lane, then the bottom towers (prevents clipping weird)
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

function resetSpriteArray(array) {
    for (var i = 0; i < array.length; i++) {
	array[i].x = -100;
	array[i].y = -100;
    }
}

function draw(units){
    // first, reset all units to base position.
    for (var i = -1; i <= 11; i++) {
	resetSpriteArray(prerenderedMobs['' + i + '-blue']);
	resetSpriteArray(prerenderedMobs['' + i + '-red']);
    }
    for (var i = 50; i <= 59; i++) {
	resetSpriteArray(prerenderedMobs['' + i + '-blue']);
	resetSpriteArray(prerenderedMobs['' + i + '-red']);
    }
    for (var i = -1; i <= 11; i++) {
	prerenderedMobIterators['' + i + '-blue'] = 0;
	prerenderedMobIterators['' + i + '-red'] = 0;
    }
    for (var i = 50; i <= 59; i++) {
	prerenderedMobIterators['' + i + '-blue'] = 0;
	prerenderedMobIterators['' + i + '-red'] = 0;
    }
    
    // draw units
    for (var i = 0; i < units.length; i++) {
	if (units[i].enum == -2) {
	    continue;
	}
	var thisUnit;
	var unitType = '' + units[i].enum + '-';
	if (units[i].owner == 1) {
	    unitType += 'blue';
	}
	else {
	    unitType += 'red';
	}

	// if we haven't prerendered enough mobs of this type, we make more and cache them for later
	while (prerenderedMobIterators[unitType] >= prerenderedMobs[unitType].length) {
	    var newMob = new Sprite(
		TextureCache[unitType]
	    );
	    if (units[i].enum == -1 && units[i].owner == 1) {
		newMob.scale.x = -1;
	    } else if (units[i].enum != -1 && units[i].owner == 2) {
	    	newMob.scale.x = -1;
	    }
	    
	    prerenderedMobs[unitType].push(newMob);
	    stage.addChild(newMob);
	}

	thisUnit = prerenderedMobs[unitType][prerenderedMobIterators[unitType]];
	prerenderedMobIterators[unitType]++;
	
	thisUnit.x = units[i].x;
	thisUnit.y = GAME_HEIGHT - units[i].y - thisUnit.height;
	/* if (thisUnit.scale.x < 0) {
	    thisUnit.scale.x = -(thisUnit.y)/(GAME_HEIGHT);
	}
	else {
	    thisUnit.scale.x = (thisUnit.y)/(GAME_HEIGHT);
	} */
	if (units[i].enum == -1) { // objective towers kind of take up the entire lane
	    thisUnit.y += thisUnit.height/4;
	}
    }
}

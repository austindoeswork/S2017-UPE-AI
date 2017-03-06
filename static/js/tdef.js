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

// Aliases
var fogOfWar = new PIXI.Graphics();
var healthBars = new PIXI.Graphics();
var p1hp = new PIXI.Text("", {font:"10px Roboto", fill:"white", align:"left"});
var p2hp = new PIXI.Text("", {font:"10px Roboto", fill:"white", align:"right"});
p1hp.x = 10;
p2hp.x = GAME_WIDTH - 10;

var p1info = new PIXI.Text("P1\nBits\nIncome", {font:"25px Roboto", fill:"#343435", align:"left"});
var p2info = new PIXI.Text("P2\nBits\nIncome", {font:"25px Roboto", fill:"#343435", align:"right"});
p1info.x = 10;
p1info.y = 30;
p2info.anchor.x = 1;
p2info.x = GAME_WIDTH - 10;
p2info.y = 30;
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
var stage = new Container();

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
    stage.addChild(fogOfWar);
    stage.addChild(healthBars);
    stage.addChild(p1info);
    stage.addChild(p2info);
    resize();
    renderer.render(stage);
    readyToDisplay = true;
}

// resize when user changes page size
window.addEventListener("resize", resize);

// called by frontend troop buttons
function buyTroop(troopEnum) {
    var radioButtons = document.getElementsByName('laneRadio');
    var laneEnum;
    for(var i = 0; i < radioButtons.length; i++){
	if(radioButtons[i].checked){
            laneEnum = radioButtons[i].value;
	    // console.log(laneEnum);
	}
    }
    input = 'b'+troopEnum +' '+laneEnum;
    // console.log(input);
    send(input);
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

var myPlayer = 0; // default 0 = spectator client, 1 = player1, 2 = player2
var myUsername, myGamename;

// TODO: handle status message at beginning of websocket (with player information), right now function assumes all messages are game board messages
function renderGrid(data) {
    frames++;
    d = JSON.parse(data);
    
    if (d.hasOwnProperty("Gamename")) {
	myPlayer = d["Player"];
	myUsername = d["Username"];
	myGamename = d["Gamename"];
	return; // don't try to render status messages
    }
    
    units = d.p1.troops.concat(d.p2.troops); // TODO: make it so the top towers are drawn first, then the lane, then the bottom towers (prevents clipping weird)
    for (i = 0; i < d.p1.towers.length; i++) {
	console.log(d.p1.towers[i].enum);
	if (d.p1.towers[i].enum != -3) { // ignore empty plots
	    units.push(d.p1.towers[i]);
	}
    }
    for (i = 0; i < d.p2.towers.length; i++) {
	if (d.p2.towers[i].enum != -3) { // ignore empty plots
	    units.push(d.p2.towers[i]);
	}
    }
    units.push(d.p1.mainCore);
    units.push(d.p2.mainCore);

    p1info.text = d.p1.name + "\nBits: " + d.p1.bits + "\nIncome: " + d.p1.income;
    p2info.text = d.p2.name + "\nBits: " + d.p2.bits + "\nIncome: " + d.p2.income;
    draw(units);

    if (myPlayer == 1) {
	fogOfWar.clear();
	fogOfWar.beginFill(0xd3d3d3);
	fogOfWar.alpha = 0.25;
	p1extra = d.p1.horizonMax - 760;
	p1scale = p1extra * 0.55;
	fogOfWar.drawPolygon([
	    800 + p1scale, 0,
	    800 + p1extra, 600,
	    1600, 600, // bottom right
	    1600, 0, //top right
	    800 + p1scale, 0,
	]);
	fogOfWar.endFill();
    } else if (myPlayer == 2) {
	fogOfWar.clear();
	fogOfWar.beginFill(0xd3d3d3);
	fogOfWar.alpha = 0.25;
	p2extra = 840 - d.p2.horizonMin;
	p2scale = p2extra * 0.55;
	fogOfWar.drawPolygon([
	    800 - p2scale, 0,
	    800 - p2extra, 600,
	    0, 600, // bottom left
	    0, 0, //top left
	    800 - p2scale, 0,
	]);
	fogOfWar.endFill();
    }

    // results screens
    if (d.p1.mainCore.hp <= 0) {
	ws.close();
	var resultScreen = new PIXI.Graphics();
	resultScreen.alpha = .5;
	resultScreen.beginFill(0x441416);
	resultScreen.drawRect(0, 0, 1600, 600);
	stage.addChild(resultScreen);
	var resultText = new PIXI.Text("RED WINS!", {font:"50px Roboto", fill:"white", align:"center"});
	resultText.anchor.x = 0.5;
	resultText.x = 800;
	resultText.anchor.y = 0.5;
	resultText.y = 300;
	stage.addChild(resultText);

	for (var i = 0; i < 12; i++) {
	    var newMob = new Sprite(
		TextureCache['' + i + '-red']
	    );
	    newMob.x = 250 + i * 100;
	    newMob.y = 400 - newMob.height;
	    stage.addChild(newMob);
	}
    }
    else if (d.p2.mainCore.hp <= 0) {
	ws.close();
	var resultScreen = new PIXI.Graphics();
	resultScreen.alpha = .5;
	resultScreen.beginFill(0x13223a);
	resultScreen.drawRect(0, 0, 1600, 600);
	stage.addChild(resultScreen);
	var resultText = new PIXI.Text("BLUE WINS!", {font:"50px Roboto", fill:"white", align:"center"});
	resultText.anchor.x = 0.5;
	resultText.x = 800;
	resultText.anchor.y = 0.5;
	resultText.y = 300;
	stage.addChild(resultText);

	for (var i = 0; i < 12; i++) {
	    var newMob = new Sprite(
		TextureCache['' + i + '-blue']
	    );
	    newMob.x = 250 + i * 100;
	    newMob.y = 400 - newMob.height;
	    stage.addChild(newMob);
	}
    }

    healthBars.clear();
    healthBars.beginFill(0x13223a); // dark blue
    healthBars.drawRect(0, 0, 600, 20);
    healthBars.beginFill(0x4286f4); // light blue
    healthBars.drawRect(0, 0, 600 * d.p1.mainCore.hp / d.p1.mainCore.maxhp, 20);
    healthBars.beginFill(0x441416);
    healthBars.drawRect(1000, 0, 600, 20);
    healthBars.beginFill(0xb2373b);
    healthBars.drawRect(1600 - 600 * d.p2.mainCore.hp / d.p2.mainCore.maxhp, 0, 600 * d.p2.mainCore.hp / d.p2.mainCore.maxhp, 20);	

    renderer.render(stage);
    // document.getElementById("fps").innerHTML = 1000/(Date.now() - timestamp);
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
	if (units[i].enum == -2) { // don't draw main core
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
	if (units[i].y >= 425) {
	    thisUnit.x = 300 + (GAME_WIDTH - 600)/GAME_WIDTH * thisUnit.x;
	}
	else if (units[i].y >= 230) {
	    thisUnit.x = 150 + (GAME_WIDTH - 300)/GAME_WIDTH * thisUnit.x;
	}
	if (units[i].enum == -1) { // objective towers kind of take up the entire lane
	    thisUnit.y += thisUnit.height/4;
	}
    }
}

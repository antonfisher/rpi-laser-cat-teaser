// create fullscreen canvas DOM element
const canvasEl = document.createElement('canvas');
canvasEl.id = 'box';
canvasEl.style = 'border:5px solid orange;cursor:none';

if (document.body.clientWidth > document.body.clientHeight) {
  canvasEl.height = document.body.clientHeight - 50;
  canvasEl.width = (canvasEl.height / 3) * 4 - 50;
} else {
  canvasEl.width = document.body.clientWidth - 50;
  canvasEl.height = (canvasEl.width / 4) * 3 - 50;
}

const logEl = document.createElement('pre');
logEl.style = 'opacity:0.3;font-family:Courier New;font-size:10px;position:absolute;top:25px;left:30px';
let l = {};
function log(val) {
  l = { ...l, ...val };
  logEl.innerHTML = Object.keys(l)
    .map(k => `${k}: ${l[k]}`)
    .join('\n');
}

// add canvas and log to the page
document.body.appendChild(canvasEl);
document.body.appendChild(logEl);

// canvas 2d context
const ctx = canvasEl.getContext('2d');
const W = canvasEl.width;
const H = canvasEl.height;
const centerX = W / 2;
const centerY = H / 2;

// laser params
const laserR = 5;
const laserSpinR = 10; // spin
let laserX = centerX;
let laserY = centerY;

// mouse events
const catR = Math.min(H, W) / 10;
const keepAwayR = catR * 5;
let catX = centerX;
let catY = centerY;
canvasEl.addEventListener('mousemove', ev => {
  catX = ev.layerX - catR / 2;
  catY = ev.layerY - catR / 2;

  //algorithmKeepTheCenter(catX, catY);
  algorithmRunAway(catX, catY);

  laserX = Math.min(W - laserSpinR / 2 - laserR / 2, Math.max(laserX, laserSpinR / 2  - laserR / 2));
  laserY = Math.min(H - laserSpinR / 2 - laserR / 2, Math.max(laserY, laserSpinR / 2  - laserR / 2));

  log({ keepAwayR, catX, catY, laserX, laserY });
});

const path = []

// laser dot runs away from the cat if it too close
function algorithmRunAway(catX, catY) {
  fromCatToLaser = distance(catX, catY, laserX, laserY);

  if (fromCatToLaser < keepAwayR) {
    laserX = catX + (keepAwayR * (laserX - catX)) / Math.sqrt(Math.pow(laserX - catX, 2) + Math.pow(laserY - catY, 2));
    laserY = catY + (keepAwayR * (laserY - catY)) / Math.sqrt(Math.pow(laserX - catX, 2) + Math.pow(laserY - catY, 2));

    // pushed out of canvas
    let alpha = (catY - laserY < 0 ? 1 : -1) * Math.acos((catX - laserX) / keepAwayR) - Math.PI / 2;
    if (laserX < 0 || laserX > W || laserY < 0 || laserY > H) {
      console.log('## alpha', alpha, alpha * 180 / Math.PI);
      log({ alpha: alpha * 180 / Math.PI });
    }
    while (laserX < 0 || laserX > W || laserY < 0 || laserY > H) {
      laserX = catX + keepAwayR * Math.sin(alpha);
      laserY = catY + keepAwayR * Math.cos(alpha);
      alpha += 0.01;

      //debug
      path.push([laserX, laserY])
    }
  }
}

// laser dot tries to keep the center or runs about the cat
function algorithmKeepTheCenter(catX, catY) {
  laserX = catX + (keepAwayR * (centerX - catX)) / Math.sqrt(Math.pow(centerX - catX, 2) + Math.pow(centerY - catY, 2));
  laserY = catY + (keepAwayR * (centerY - catY)) / Math.sqrt(Math.pow(centerX - catX, 2) + Math.pow(centerY - catY, 2));

  fromCatToCenter = distance(catX, catY, centerX, centerY);
  if (fromCatToCenter > keepAwayR) {
    laserX = centerX;
    laserY = centerY;
  }
}

// laser dot bounces in the canvas
function algorithmBounce() {
  const speed = Math.min(H, W) / 200;
  let laserXSpeed = speed;
  let laserYSpeed = speed;
  setInterval(() => {
    const d = distance(catX, catY, laserX, laserY);
    log({ d });
    if (laserX >= W || laserX <= 0) {
      laserXSpeed *= -1;
    }
    if (laserY >= H || laserY <= 0) {
      laserYSpeed *= -1;
    }
    laserX = laserX + laserXSpeed;
    laserY = laserY + laserYSpeed;
  }, 10);
}
//algorithmBounce()

function distance(x1, y1, x2, y2) {
  return Math.sqrt(Math.pow(x1 - x2, 2) + Math.pow(y1 - y2, 2));
}

// spin laser dot
let laserShiftSpeed = 0.1;
let laserShiftAngle = 0;
let laserShiftX = 0;
let laserShiftY = 0;
setInterval(() => {
  laserShiftX = laserSpinR * Math.sin(laserShiftAngle);
  laserShiftY = laserSpinR * Math.cos(laserShiftAngle);
  laserShiftAngle += laserShiftSpeed;
  if (laserShiftAngle <= 0 || laserShiftAngle >= 2 * Math.PI) {
    laserShiftSpeed *= -1;
  }
}, 20);

// the main loop
function loop() {
  ctx.clearRect(0, 0, canvasEl.width, canvasEl.height);

  path.forEach((v) => {
    ctx.beginPath();
      ctx.arc(v[0], v[1], 10, 0, Math.PI * 2, true);
      ctx.strokeStyle = '#dfd';
      ctx.stroke();
      ctx.closePath();
  });

  // cat
  ctx.beginPath();
  ctx.arc(catX, catY, catR, 0, Math.PI * 2, true);
  ctx.moveTo(catX, catY);
  ctx.arc(catX, catY, 1, 0, Math.PI * 2, true);
  ctx.strokeStyle = '#00f';
  ctx.fillStyle = '#99f';
  ctx.fill();
  ctx.stroke();
  ctx.closePath();

  // cat - left ear
  ctx.beginPath();
  ctx.moveTo(catX - catR * 0.8, catY - catR * 0.65);
  ctx.lineTo(catX - catR * 0.95, catY - catR * 0.95);
  ctx.lineTo(catX - catR * 0.65, catY - catR * 0.8);
  ctx.strokeStyle = '#00f';
  ctx.fillStyle = '#99f';
  ctx.fill();
  ctx.stroke();
  ctx.closePath();

  // cat - right ear
  ctx.beginPath();
  ctx.moveTo(catX + catR * 0.8, catY - catR * 0.65);
  ctx.lineTo(catX + catR * 0.95, catY - catR * 0.95);
  ctx.lineTo(catX + catR * 0.65, catY - catR * 0.8);
  ctx.strokeStyle = '#00f';
  ctx.fillStyle = '#99f';
  ctx.fill();
  ctx.stroke();
  ctx.closePath();

  // keep away radius
  ctx.beginPath();
  ctx.arc(catX, catY, keepAwayR, 0, Math.PI * 2, true);
  ctx.strokeStyle = '#eef';
  ctx.setLineDash([4, 2]);
  ctx.stroke();
  ctx.setLineDash([]);
  ctx.closePath();

  // laser dot
  ctx.beginPath();
  ctx.arc(laserX + laserShiftX, laserY + laserShiftY, laserR, 0, Math.PI * 2, true);
  ctx.strokeStyle = '#f00';
  ctx.fillStyle = '#f99';
  ctx.fill();
  ctx.stroke();
  ctx.closePath();

  //center
  ctx.beginPath();
  ctx.moveTo(centerX - 5, centerY);
  ctx.lineTo(centerX + 5, centerY);
  ctx.moveTo(centerX, centerY - 5);
  ctx.lineTo(centerX, centerY + 5);
  ctx.strokeStyle = 'rgba(0,0,0, 0.3)';
  ctx.stroke();
  ctx.closePath();

  setTimeout(loop, 10);
}

// run program
loop();

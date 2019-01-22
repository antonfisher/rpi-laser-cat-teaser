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
const laserSpinR = 5; // spin
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

  laserX = Math.min(W - laserSpinR / 2 - laserR / 2, Math.max(laserX, laserSpinR / 2 - laserR / 2));
  laserY = Math.min(H - laserSpinR / 2 - laserR / 2, Math.max(laserY, laserSpinR / 2 - laserR / 2));

  log({ keepAwayR, catX, catY, laserX, laserY });
});

// laser dot runs away from the cat if it too close
function algorithmRunAway(catX, catY) {
  // closest point from the current laser position to the "keep away" circle
  const kaX = catX + (keepAwayR * (laserX - catX)) / Math.sqrt(Math.pow(laserX - catX, 2) + Math.pow(laserY - catY, 2));
  const kaY = catY + (keepAwayR * (laserY - catY)) / Math.sqrt(Math.pow(laserX - catX, 2) + Math.pow(laserY - catY, 2));

  if (distance(catX, catY, laserX, laserY) < distance(catX, catY, kaX, kaY)) {
    laserX = kaX;
    laserY = kaY;
  }

  // pushed out of canvas
  if (laserX < 0 || laserX > W || laserY < 0 || laserY > H) {
    const intersections = [];
    // pushed to the top
    if (catY - keepAwayR < 0) {
      const dx = Math.sqrt(Math.pow(keepAwayR, 2) - Math.pow(catY, 2));
      if (0 <= catX + dx && catX + dx <= W) {
        intersections.push({ x: catX + dx, y: 0 });
      }
      if (0 <= catX - dx && catX - dx <= W) {
        intersections.push({ x: catX - dx, y: 0 });
      }
    }
    // pushed to the bottom
    if (catY + keepAwayR > H) {
      const dx = Math.sqrt(Math.pow(keepAwayR, 2) - Math.pow(H - catY, 2));
      if (0 <= catX + dx && catX + dx <= W) {
        intersections.push({ x: catX + dx, y: H });
      }
      if (0 <= catX - dx && catX - dx <= W) {
        intersections.push({ x: catX - dx, y: H });
      }
    }
    // pushed to the left
    if (catX - keepAwayR < 0) {
      const dy = Math.sqrt(Math.pow(keepAwayR, 2) - Math.pow(catX, 2));
      if (0 <= catY + dy && catY + dy <= H) {
        intersections.push({ x: 0, y: catY + dy });
      }
      if (0 <= catY - dy && catY - dy <= H) {
        intersections.push({ x: 0, y: catY - dy });
      }
    }
    // pushed to the right
    if (catX + keepAwayR > W) {
      const dy = Math.sqrt(Math.pow(keepAwayR, 2) - Math.pow(W - catX, 2));
      if (0 <= catY + dy && catY + dy <= H) {
        intersections.push({ x: W, y: catY + dy });
      }
      if (0 <= catY - dy && catY - dy <= H) {
        intersections.push({ x: W, y: catY - dy });
      }
    }
    log({ intersections_length: intersections.length });
    if (intersections.length > 0) {
      let minDistance = keepAwayR;
      let minDistanceX = intersections[0].x;
      let minDistanceY = intersections[0].y;
      for (let i = 0; i < intersections.length; i++) {
        const d = distance(laserX, laserY, intersections[i].x, intersections[i].y);
        if (d < minDistance) {
          minDistance = d;
          minDistanceX = intersections[i].x;
          minDistanceY = intersections[i].y;
        }
      }
      laserX = minDistanceX;
      laserY = minDistanceY;
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

function distance(x0, y0, x1, y1) {
  return Math.sqrt(Math.pow(x0 - x1, 2) + Math.pow(y0 - y1, 2));
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
}, 10);

// the main loop
function loop() {
  ctx.clearRect(0, 0, canvasEl.width, canvasEl.height);

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

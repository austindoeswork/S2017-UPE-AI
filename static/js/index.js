'use strict';

const QUALS_DATE = new Date(2017, 2, 29, 23, 59, 59, 59);
const DAYS_IN_MILLISECONDS = 1000 * 60 * 60 * 24;
const HOURS_IN_MILLISECONDS = 1000 * 60 * 60;
const MINUTES_IN_MILLISECONDS = 1000 * 60;

function updateTime() {
  var $countdown = $('#countdown');
  var now = new Date();
  var time_remaining = QUALS_DATE - now;
  var days = Math.floor(time_remaining / DAYS_IN_MILLISECONDS);
  time_remaining = time_remaining % DAYS_IN_MILLISECONDS;
  var hours = Math.floor(time_remaining / HOURS_IN_MILLISECONDS);
  time_remaining = time_remaining % HOURS_IN_MILLISECONDS;
  var minutes = Math.floor(time_remaining / MINUTES_IN_MILLISECONDS);
  time_remaining = time_remaining % MINUTES_IN_MILLISECONDS;
  var seconds = Math.floor(time_remaining / 1000);
  $countdown.text(`${days}:${hours}:${minutes}:${seconds}`);
}

window.setInterval(updateTime, 1000);

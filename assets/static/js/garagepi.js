"use strict";

$(document).ready(function(){

  var $btnLight = $("#btnLight");

  var lightText = $btnLight.text();
  var lightOn;
  if (lightText = "Light On") {
    lightOn = true;
  } else {
    lightOn = false;
  }

  function toggleGarageDoor() {
    $.post("/toggle");
  }

  function turnLightOn() {
    $.post("/light?state=on");
  }

  function turnLightOff() {
    $.post("/light?state=off");
  }

  $("#btnDoorToggle").on("click", function( event ) {
    toggleGarageDoor();
  });

  $btnLight.on("click", function( event ) {
    if (lightOn) {
      turnLightOff();
      lightOn = false;
      $btnLight.text("Light On")
    } else {
      turnLightOn();
      lightOn = true;
      $btnLight.text("Light Off")
    }
  });
});

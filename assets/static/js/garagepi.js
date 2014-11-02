"use strict";

$(document).ready(function(){

  var $btnLight = $("#btnLight");

  var lightText = $btnLight.text();
  var lightOn = (lightText == "Light Off");

  function toggleGarageDoor() {
    $.post("/toggle");
  }

  function turnLightOn() {
    $.post("/light?state=on", parseLightState);
  }

  function turnLightOff() {
    $.post("/light?state=off", parseLightState);
  }

  function parseLightState(data) {
    data = $.parseJSON(data);
    lightOn = data.LightOn;

    if (lightOn) {
      $btnLight.text("Light Off");
    } else {
      $btnLight.text("Light On");
    }

    if (!data.StateKnown) {
      $btnLight.prop('disabled', true);
    }
  }

  $("#btnDoorToggle").on("click", function( event ) {
    toggleGarageDoor();
  });

  $btnLight.on("click", function( event ) {
    if (lightOn) {
      turnLightOff();
    } else {
      turnLightOn();
    }
  });
});

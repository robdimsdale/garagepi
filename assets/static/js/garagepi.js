"use strict";

$(document).ready(function(){

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
    toggleGarageDoor()
  });

  $("#btnLightOn").on("click", function( event ) {
    turnLightOn()
  });

  $("#btnLightOff").on("click", function( event ) {
    turnLightOff()
  });
});

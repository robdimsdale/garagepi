"use strict";

$(document).ready(function(){

  function turnLightOn() {
  $.post("/light?state=on");
  }

  function turnLightOff() {
  $.post("/light?state=off");
  }

  $("#btnLightOn").on("click", function( event ) {
    turnLightOn()
  });

  $("#btnLightOff").on("click", function( event ) {
    turnLightOff()
  });
});

var modalText = document.querySelector("#modal-text")
var modal = document.getElementById('myModal');
var modalContent = document.querySelector(".modal-content")

// Get the <span> element that closes the modal
var span = document.getElementsByClassName("close")[0];

// When the user clicks on <span> (x), close the modal
span.onclick = function() {
  // modal.style.display = "none";
  modal.style.visibility = "hidden";
  modal.style.opacity = "0";
  // modalContent.style.visibility = "hidden"
  // modalContent.style.opacity = "0"
}

// When the user clicks anywhere outside of the modal, close it
window.onclick = function(event) {
  if (event.target == modal) {
    // modal.style.display = "none";
    modal.style.visibility = "hidden";
    modal.style.opacity = "0";
    // modalContent.style.visibility = "hidden"
    // modalContent.style.opacity = "0"
  }
}

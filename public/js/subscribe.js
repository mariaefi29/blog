var form = document.querySelector(".sing-up-form");
var noshow = document.getElementById("r0d0");
noshow.value = 454;
form.addEventListener('submit', function(e){
  var email = document.querySelector(".email");
  var xhr = new XMLHttpRequest();
  xhr.open('POST', '/subscribe', true);
  xhr.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
  xhr.addEventListener('readystatechange', function() {
      if(xhr.readyState === XMLHttpRequest.DONE && xhr.status === 200){
          form.reset();
          modal.style.visibility = "visible";
          modal.style.opacity = "1";
          modalText.textContent = xhr.responseText;
      }
  });
  e.preventDefault();
  xhr.send("email="+email.value+"&noshow="+noshow.value);
})

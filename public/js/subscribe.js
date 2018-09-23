var form = document.querySelector(".sing-up-form");
console.log(form);
form.addEventListener('submit', function(e){
  var email = document.querySelector(".email");
  console.log(email.value);
  var xhr = new XMLHttpRequest();
  xhr.open('POST', '/subscribe', true);
  xhr.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
  xhr.addEventListener('readystatechange', function() {
      if(xhr.readyState === XMLHttpRequest.DONE && xhr.status === 200){
          form.reset();
          // modal.style.display = "block";
          modal.style.visibility = "visible";
          modal.style.opacity = "1";
          modalText.textContent = xhr.responseText;
      }
  });
  e.preventDefault();
  xhr.send("email="+email.value);
})

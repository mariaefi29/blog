var likeButton = document.querySelector(".like-button");
console.log(likeButton);
likeButton.onclick = function(e) {
  var xhr = new XMLHttpRequest();
  xhr.open('POST', '/posts/show/{{.IDstr}}', true);
  xhr.addEventListener('readystatechange', function() {
      if(xhr.readyState === XMLHttpRequest.DONE && xhr.status === 200){
          alert(xhr.responseText);
      }
  });
  e.preventDefault();
  xhr.send();
};

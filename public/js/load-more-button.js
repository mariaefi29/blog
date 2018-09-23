
var parent = document.querySelector('.posts'),
    items  = parent.querySelectorAll('.post'),
    loadMoreBtn =  document.querySelector('#load-more-btn'),
    maxItems = 3,
    hiddenClass = "visually-hidden";

if (items.length < 3) {loadMoreBtn.style.display = 'none';}

[].forEach.call(items, function(item, idx){
    if (idx > maxItems - 1) {
        item.classList.add(hiddenClass);
    }
});

loadMoreBtn.addEventListener('click', function(){

  [].forEach.call(document.querySelectorAll('.' + hiddenClass), function(item, idx){
      console.log(item);
      if (idx <= maxItems - 1) {
          item.classList.remove(hiddenClass);
          console.log(item);
      }

      if ( document.querySelectorAll('.' + hiddenClass).length === 0) {
          loadMoreBtn.style.display = 'none';
      }

  });

});

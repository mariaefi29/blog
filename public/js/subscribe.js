var form = document.querySelector(".sing-up-form");
var noshow = document.getElementById("r0d0");
noshow.value = 454;
form.addEventListener('submit', function(e){
  e.preventDefault();
    var token = window.grecaptcha.getResponse(recaptchaId);
    // if no token, mean user is not validated yet
    if (!token) {
        window.grecaptcha.reset(recaptchaId)
        window.grecaptcha.execute(recaptchaId);
    }
});

window.onScriptLoad = function () {
    // this callback will be called by recaptcah/api.js once its loaded. If we used
    // render=explicit as param in script src, then we can explicitly render reCaptcha at this point

    // element to "render" invisible captcha in
    var htmlEl = document.querySelector('.g-recaptcha');

    // option to captcha
    var captchaOptions = {
        sitekey: '6LcKavQUAAAAAG1pch276IT2nj2ulSvLF5RP8bhW',
        size: 'invisible',
        // tell reCaptcha which callback to notify when user is successfully verified.
        // if this value is string, then it must be name of function accessible via window['nameOfFunc'],
        // and passing string is equivalent to specifying data-callback='nameOfFunc', but it can be
        // reference to an actual function
        callback: window.onUserVerified
    };

    // now render
    recaptchaId = window.grecaptcha.render(htmlEl, captchaOptions, true);
};

// this is assigned from "data-callback" or render()'s "options.callback"
window.onUserVerified = function (token) {
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

    xhr.send("email="+email.value+"&noshow="+noshow.value+"&g-recaptcha-response="+token);
};
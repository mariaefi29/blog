(function () {
  var measurementId = window.blogGoogleAnalyticsId;
  var storageKey = "marialife.analyticsConsent";

  if (!measurementId) {
    return;
  }

  function getConsent() {
    try {
      return window.localStorage.getItem(storageKey);
    } catch (err) {
      return null;
    }
  }

  function setConsent(value) {
    try {
      window.localStorage.setItem(storageKey, value);
    } catch (err) {
      // If storage is blocked, keep the choice for the current page only.
      window.blogAnalyticsConsent = value;
    }
  }

  function removeElement(selector) {
    var element = document.querySelector(selector);
    if (element) {
      element.parentNode.removeChild(element);
    }
  }

  function consentDefaults() {
    return {
      analytics_storage: "denied",
      ad_storage: "denied",
      ad_user_data: "denied",
      ad_personalization: "denied",
      wait_for_update: 500
    };
  }

  function deniedConsent() {
    return {
      analytics_storage: "denied",
      ad_storage: "denied",
      ad_user_data: "denied",
      ad_personalization: "denied"
    };
  }

  function grantedAnalyticsConsent() {
    return {
      analytics_storage: "granted",
      ad_storage: "denied",
      ad_user_data: "denied",
      ad_personalization: "denied"
    };
  }

  function gtag() {
    window.dataLayer.push(arguments);
  }

  function loadAnalytics() {
    if (window.blogGoogleAnalyticsLoaded) {
      window.gtag("consent", "update", grantedAnalyticsConsent());
      return;
    }

    window.blogGoogleAnalyticsLoaded = true;
    window.dataLayer = window.dataLayer || [];
    window.gtag = window.gtag || gtag;
    window.gtag("consent", "default", consentDefaults());

    var script = document.createElement("script");
    script.async = true;
    script.src = "https://www.googletagmanager.com/gtag/js?id=" + encodeURIComponent(measurementId);
    document.head.appendChild(script);

    window.gtag("consent", "update", grantedAnalyticsConsent());
    window.gtag("js", new Date());
    window.gtag("config", measurementId, { anonymize_ip: true });
  }

  function expireCookie(name, domain) {
    var cookie = name + "=; expires=Thu, 01 Jan 1970 00:00:00 GMT; path=/; SameSite=Lax";
    document.cookie = domain ? cookie + "; domain=" + domain : cookie;
  }

  function cookieDomains() {
    var host = window.location.hostname;
    var domains = [host];
    var parts = host.split(".");

    for (var i = 0; i < parts.length - 1; i++) {
      domains.push("." + parts.slice(i).join("."));
    }

    return domains;
  }

  function deleteAnalyticsCookies() {
    var domains = cookieDomains();
    var cookies = document.cookie ? document.cookie.split(";") : [];

    cookies.forEach(function (cookie) {
      var name = cookie.split("=")[0].trim();
      var isAnalyticsCookie = name === "_ga" || name.indexOf("_ga_") === 0 || name === "_gid" || name.indexOf("_gat") === 0;

      if (!isAnalyticsCookie) {
        return;
      }

      expireCookie(name);
      domains.forEach(function (domain) {
        expireCookie(name, domain);
      });
    });
  }

  function rejectAnalytics() {
    if (window.gtag) {
      window.gtag("consent", "update", deniedConsent());
    }

    deleteAnalyticsCookies();
  }

  function closeBanner() {
    removeElement(".cookie-consent");
    removeElement(".cookie-consent-backdrop");
  }

  function renderBanner() {
    removeElement(".cookie-consent");
    removeElement(".cookie-consent-backdrop");

    var backdrop = document.createElement("div");
    backdrop.className = "cookie-consent-backdrop";
    backdrop.setAttribute("aria-hidden", "true");

    var banner = document.createElement("section");
    banner.className = "cookie-consent";
    banner.setAttribute("role", "dialog");
    banner.setAttribute("aria-labelledby", "cookie-consent-title");
    banner.setAttribute("aria-describedby", "cookie-consent-description");

    banner.innerHTML = [
      '<div class="cookie-consent__content">',
      '<h2 id="cookie-consent-title" class="cookie-consent__title">Let us know you agree to Google Analytics cookies.</h2>',
      '<p id="cookie-consent-description" class="cookie-consent__text">We use Google Analytics cookies only with your consent to understand how the site is used. You can reject them and continue using the site normally.</p>',
      "</div>",
      '<div class="cookie-consent__actions">',
      '<button type="button" class="cookie-consent__button cookie-consent__reject">I do not agree</button>',
      '<button type="button" class="cookie-consent__button cookie-consent__accept">I agree</button>',
      "</div>"
    ].join("");

    banner.querySelector(".cookie-consent__reject").addEventListener("click", function () {
      setConsent("denied");
      rejectAnalytics();
      closeBanner();
    });

    banner.querySelector(".cookie-consent__accept").addEventListener("click", function () {
      setConsent("granted");
      loadAnalytics();
      closeBanner();
    });

    document.body.appendChild(backdrop);
    document.body.appendChild(banner);
  }

  function init() {
    var consent = getConsent() || window.blogAnalyticsConsent;

    if (consent === "granted") {
      loadAnalytics();
      return;
    }

    if (consent === "denied") {
      rejectAnalytics();
      return;
    }

    renderBanner();
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();

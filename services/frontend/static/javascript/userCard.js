import * as util from '/static/javascript/util.js';
"use strict";

export function setUserCardDetails(user) {
    document.getElementById("userUsername").textContent = user.accdetails.personaname;
    document.getElementById("userCountry").textContent = util.countryCodeToName(user.accdetails.loccountrycode) === "" ? 'unknown' : util.countryCodeToName(user.accdetails.loccountrycode);
    document.getElementById("userFriendCount").textContent = user.friendids.length;
    
    const creationDate = new Date(user.accdetails.timecreated*1000);
    const dateString = `${creationDate.getDate()} ${creationDate.toLocaleString('default', { month: 'long' })} ${creationDate.getFullYear()}`;
    const timeSinceString = `(${util.timezSince(creationDate)})`
    document.getElementById("userCreationDate").textContent = `${dateString} ${timeSinceString}`;
    
    document.getElementById("userProfile").innerHTML = `<a href="${user.accdetails.profileurl}">Profile Link</a>`;
    document.getElementById("userAvatar").src = user.accdetails.avatar.split(".jpg").join("") + "_full.jpg";

    document.getElementById("userUsername").classList.remove("skeleton");
    document.getElementById("userCountry").classList.remove("skeleton");
    document.getElementById("userFriendCount").classList.remove("skeleton");
    document.getElementById("userCreationDate").classList.remove("skeleton");
    document.getElementById("userProfile").classList.remove("skeleton");
    document.getElementById("userAvatar").classList.remove("skeleton");
}

export function setCrawlPageUserCardDetails(user, idPrefix) {
  document.getElementById(`${idPrefix}UserUsername`).textContent = user.accdetails.personaname;
  document.getElementById(`${idPrefix}UserCountry`).textContent = util.countryCodeToName(user.accdetails.loccountrycode) === "" ? 'unknown' : util.countryCodeToName(user.accdetails.loccountrycode);
  document.getElementById(`${idPrefix}UserFriendCount`).textContent = user.friendids.length;
  
  const creationDate = new Date(user.accdetails.timecreated*1000);
  const dateString = `${creationDate.getDate()} ${creationDate.toLocaleString('default', { month: 'long' })} ${creationDate.getFullYear()}`;
  const timeSinceString = `(${util.timezSince(creationDate)})`
  document.getElementById(`${idPrefix}UserCreationDate`).textContent = `${dateString} ${timeSinceString}`;
  
  document.getElementById(`${idPrefix}UserProfile`).innerHTML = `<a href="${user.accdetails.profileurl}">Profile Link</a>`;
  document.getElementById(`${idPrefix}UserAvatar`).src = user.accdetails.avatar.split(".jpg").join("") + "_full.jpg";

  document.getElementById(`${idPrefix}UserUsername`).classList.remove("skeleton");
  document.getElementById(`${idPrefix}UserCountry`).classList.remove("skeleton");
  document.getElementById(`${idPrefix}UserFriendCount`).classList.remove("skeleton");
  document.getElementById(`${idPrefix}UserCreationDate`).classList.remove("skeleton");
  document.getElementById(`${idPrefix}UserProfile`).classList.remove("skeleton");
  document.getElementById(`${idPrefix}UserAvatar`).classList.remove("skeleton");
}

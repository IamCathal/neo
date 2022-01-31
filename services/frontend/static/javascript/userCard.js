export function setUserCardDetails(userObj) {
    const user = userObj.User;
    console.log(user)
    document.getElementById("userUsername").textContent = user.accdetails.personaname;
    document.getElementById("userRealName").textContent = "idk";
    document.getElementById("userFriendCount").textContent = user.friendids.length;
    
    const creationDate = new Date(user.accdetails.timecreated*1000);
    const dateString = `${creationDate.getDate()} ${creationDate.toLocaleString('default', { month: 'long' })} ${creationDate.getFullYear()}`;
    const timeSinceString = `(${timezSince(creationDate)} ago)`
    document.getElementById("userCreationDate").textContent = `${dateString} ${timeSinceString}`;
    
    document.getElementById("userSteamID").textContent = user.accdetails.steamid;
    document.getElementById("userAvatar").src = user.accdetails.avatar.split(".jpg").join("") + "_full.jpg";

    document.getElementById("userUsername").classList.remove("skeleton");
    document.getElementById("userRealName").classList.remove("skeleton");
    document.getElementById("userFriendCount").classList.remove("skeleton");
    document.getElementById("userCreationDate").classList.remove("skeleton");
    document.getElementById("userSteamID").classList.remove("skeleton");
    document.getElementById("userAvatar").classList.remove("skeleton");
}

// COMMON
export function timezSince(targetDate) {
    let seconds = Math.floor((new Date()-targetDate)/1000)
    let interval = seconds / 31536000 
    if (interval > 1) {
        return Math.floor(interval) + " years";
    }
    interval = seconds / 2592000; // months
    if (interval > 1) {
        return Math.floor(interval) + " months";
      }
    interval = seconds / 86400; // days
    if (interval > 1) {
      return Math.floor(interval) + "d ago";
    }
    interval = seconds / 3600;
    if (interval > 1) {
      return Math.floor(interval) + "h ago";
    }
    interval = seconds / 60;
    if (interval > 1) {
      return Math.floor(interval) + "m ago";
    }
    return Math.floor(seconds) + "s";
}
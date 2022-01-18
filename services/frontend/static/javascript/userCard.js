function setUserCardDetails(userObj) {
    console.log(userObj)
    document.getElementById("userUsername").textContent = userObj.user.accdetails.personaname;
    document.getElementById("userRealName").textContent = "idk";
    document.getElementById("userFriendCount").textContent = userObj.user.friendids.length;
    
    const creationDate = new Date(userObj.user.accdetails.timecreated*1000);
    const dateString = `${creationDate.getDate()} ${creationDate.toLocaleString('default', { month: 'long' })} ${creationDate.getFullYear()}`;
    const timeSinceString = `(${timezSince(creationDate)} ago)`
    document.getElementById("userCreationDate").textContent = `${dateString} ${timeSinceString}`;
    
    document.getElementById("userSteamID").textContent = userObj.user.accdetails.steamid;
    document.getElementById("userAvatar").src = userObj.user.accdetails.avatar.split(".jpg").join("") + "_full.jpg";

    document.getElementById("userUsername").classList.remove("skeleton");
    document.getElementById("userRealName").classList.remove("skeleton");
    document.getElementById("userFriendCount").classList.remove("skeleton");
    document.getElementById("userCreationDate").classList.remove("skeleton");
    document.getElementById("userSteamID").classList.remove("skeleton");
    document.getElementById("userAvatar").classList.remove("skeleton");
}
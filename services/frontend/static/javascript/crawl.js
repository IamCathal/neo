function isValidFormatSteamID(steamID) {
    if (steamID.length != 17) {
        return false
    }
    const regex = /([0-9]){17}/g
    result = steamID.match(regex);
    if (result.length == 1) {
        return true
    }
    return false
}

function displayErrorForInvalidSteamID(isFirstSteamID, errorMessage) {
    console.log("display error")
    let firstSteamIDInput = document.getElementById("firstSteamID");
    let secondSteamIDInput = document.getElementById("secondSteamID");

    if (isFirstSteamID == true) {
        firstSteamIDInput.style.border = "3px solid red"
        document.getElementById("steamIDChoiceFirstErrorBox").style.display = "block";
        document.getElementById("steamIDChoiceFirstErrorBoxMessage").textContent = errorMessage;
    } else {
        secondSteamIDInput.style.border = "3px solid red"
        document.getElementById("steamIDChoiceSecondErrorBox").style.display = "block";
        document.getElementById("steamIDChoiceSecondErrorBoxMessage").textContent = errorMessage;
    }
}

function hideSteamIDInputErrors() {
    console.log("hiding errors")

    let firstSteamIDInput = document.getElementById("firstSteamID");
    let secondSteamIDInput = document.getElementById("secondSteamID");
    firstSteamIDInput.style.border = "1px solid white";
    secondSteamIDInput.style.border = "1px solid white"
    document.getElementById("steamIDChoiceFirstErrorBox").style.display = "none";
    document.getElementById("steamIDChoiceFirstErrorBoxMessage").textContent = "";

    document.getElementById("steamIDChoiceSecondErrorBox").style.display = "none";
    document.getElementById("steamIDChoiceSecondErrorBoxMessage").textContent = "";
}

function hideCrawlLoadingElements() {
    document.getElementById("crawlConfigLoadingElement").style.visibility = "hidden";
    document.getElementById("crawlConfigInnerBox").style.webkitFilter = "blur(0px)";
}

function isPublicProfile(steamID) {
    return new Promise((resolve, reject) => {
        fetch(`http://localhost:2570/isprivateprofile/${steamID}`, {
        method: "GET",
        headers: {
            "Content-Type": "application/json"
        }
    }).then(res => res.json())
    .then(data => {
        console.log(data);
        if (data.message === "public") {
            resolve(true)
        }
        resolve(false)
    }).catch(err => {
        console.error(err)
        reject(err)
        })
    })
}

function hasBeenCrawled(steamID, level) {
    return new Promise((resolve, reject) => {
        reqBody = {
            "level": parseInt(level),
            "steamid": steamID
        }
        fetch(`http://localhost:2590/api/hasbeencrawledbefore`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify(reqBody),
        }).then(res => res.json())
        .then(data => {
            console.log(data)
            resolve(data.message)
        }).catch(err => {
            reject(err)
        })
    })
}
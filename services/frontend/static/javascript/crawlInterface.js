import * as utilRequest from '/static/javascript/utilRequests.js';

let singleUserMode = true;

document.getElementById("oneUserButton").addEventListener("click", function() {
    const oneUserButton = document.getElementById("oneUserButton");
    const twoUsersButton = document.getElementById("twoUsersButton");

    document.getElementById("steamIDChoiceSecondErrorBox").style.display = "none";

    // If it's not already active
    if (!oneUserButton.classList.contains("btn-light")) {
        oneUserButton.classList.add("btn-light");
        oneUserButton.classList.remove("btn-outline-light");

        twoUsersButton.classList.add("btn-outline-light");
        twoUsersButton.classList.remove("btn-light");

        document.getElementById("secondUserInputBox").style.animation = "contract .4s ease-in-out";

        setTimeout(function() {
            document.getElementById("secondUserInputBox").style.visibility = "hidden";
        }, 400);
        singleUserMode = true;
    }
});

document.getElementById("twoUsersButton").addEventListener("click", function() {
    const oneUserButton = document.getElementById("oneUserButton");
    const twoUsersButton = document.getElementById("twoUsersButton");

    // If it's not already active
    if (!twoUsersButton.classList.contains("btn-light")) {
        twoUsersButton.classList.add("btn-light");
        twoUsersButton.classList.remove("btn-outline-light");

        oneUserButton.classList.add("btn-outline-light");
        oneUserButton.classList.remove("btn-light");

        document.getElementById("secondUserInputBox").style.visibility = "visible";
        document.getElementById("secondUserInputBox").style.animation = "expand .2s ease-in-out";
        singleUserMode = false;
    }
});

document.getElementById("levelChoice").addEventListener("input", function(event) {
    document.getElementById("chosenLevel").textContent = this.value;

    const levelHelpTextMessages = [
        "Crawl just the given user(s)",
        "Crawl the given user(s) and all of their immediate friends",
        "Crawl the given user(s), all of their immediate friends and all friends of these friends"
    ]
    const levelExampleImages = [
        "https://i.imgur.com/eoSRDK6.png",
        "https://i.imgur.com/6LbbJoV.png",
        "https://i.imgur.com/blTUZPU.png"
    ]

    document.getElementById("levelChoiceInfoBox").style.display = "block";
    document.getElementById("levelHelpText").textContent = levelHelpTextMessages[this.value-1];
    document.getElementById("levelHelpImage").src = levelExampleImages[this.value-1];
});

document.getElementById("crawlButton").addEventListener("click", function(event) {
    document.getElementById("crawlConfigLoadingElement").style.visibility = "visible";
    document.getElementById("crawlConfigInnerBox").style.webkitFilter = "blur(4px)";

    hideSteamIDInputErrors()

    const firstSteamID = document.getElementById("firstSteamID").value;
    const secondSteamID = document.getElementById("secondSteamID").value;
    const level = document.getElementById("levelChoice").value;
    console.log(firstSteamID, secondSteamID, level, singleUserMode)

    if (singleUserMode == true) {
        if (!isValidFormatSteamID(firstSteamID)) {
            displayErrorForInvalidSteamID(true, "Invalid format steamID given");
            hideCrawlLoadingElements()
            return
        }
    } else {
        if (!isValidFormatSteamID(firstSteamID)) {
            displayErrorForInvalidSteamID(true, "Invalid format steamID given");
            hideCrawlLoadingElements()
        }
        if (!isValidFormatSteamID(secondSteamID)) {
            displayErrorForInvalidSteamID(false, "Invalid format steamID given");
            hideCrawlLoadingElements()
            return
        }
    }
    console.log(singleUserMode)
    if (singleUserMode == true) {
        console.log("single user mode")
        utilRequest.isPublicProfile(firstSteamID).then((isPublic) => {
            if (!isPublic) {
                hideCrawlLoadingElements()
                displayErrorForInvalidSteamID(true, "Profile is private");
                return
            }
            document.getElementById("isPrivateCheckMark").style.filter = "invert(78%) sepia(41%) saturate(7094%) hue-rotate(81deg) brightness(111%) contrast(109%)"
            document.getElementById("crawlingInterfaceStatusText").textContent = "Checking for existing crawls"
            utilRequest.hasBeenCrawled(firstSteamID, level).then((crawlID) => {
                if (crawlID === "") {
                    document.getElementById("isCrawledBeforeCheckMark").style.filter = "invert(78%) sepia(41%) saturate(7094%) hue-rotate(81deg) brightness(111%) contrast(109%)"
                    
                    // New crawl
                    let crawlDTO = {
                        level: parseInt(level)
                    }
                    if (secondSteamID) {
                        crawlDTO["steamids"] = [firstSteamID, secondSteamID]
                    } else {
                        crawlDTO["steamids"] = [firstSteamID]
                    }
    
                    utilRequest.startCrawl(crawlDTO).then(crawlIDs => {
                        window.location.href = `/crawl?firstcrawlid=${crawlIDs[0]}`
                    }, (err) => {
                        console.error(`startCrawl: ${err}`);
                    })
                } else {
                    // Crawl already exists, reroute to that page
                    window.location.href = `/graph/${crawlID}`;
                }
                hideCrawlLoadingElements()
            }, (err) => {
                console.log(`err from hasBeenCrawl: ${err}`)
            })
        }, (err) => {
            console.log(`err from isPublic: ${err}`)
        })
    } else {
        console.log("double userrrr")
        let firstProfileIsPublic = utilRequest.isPublicProfile(firstSteamID);
        let secondProfileIsPublic = utilRequest.isPublicProfile(secondSteamID)

        Promise.all([firstProfileIsPublic, secondProfileIsPublic]).then(isPublicArr => {
            console.log(isPublicArr)
            if (isPublicArr[0] != true && isPublicArr[1] != true) {
                hideCrawlLoadingElements()
                displayErrorForInvalidSteamID(true, "Both profiles are not public");
                return
            }

            document.getElementById("isPrivateCheckMark").style.filter = "invert(78%) sepia(41%) saturate(7094%) hue-rotate(81deg) brightness(111%) contrast(109%)"
            let firstUserHasBeenCrawled = utilRequest.hasBeenCrawled(firstSteamID, level);
            let secondUserHasBeenCrawled = utilRequest.hasBeenCrawled(secondSteamID, level)

            Promise.all([firstUserHasBeenCrawled, secondUserHasBeenCrawled]).then(crawlIDs => {
                if (crawlIDs[0] != "" && crawlIDs[1] != "") {
                    // both users have been crawled before
                    console.log(`both ids are not empty ${crawlIDs}`)
                    window.location.href = `/shortestdistance?firstcrawlid=${crawlIDs[0]}&secondcrawlid=${crawlIDs[1]}`
                }

                let crawlDTO = {
                    level: parseInt(level),
                    steamids: []
                }

                if (crawlIDs[0] == "") {
                    document.getElementById("isCrawledBeforeCheckMark").style.filter = "invert(78%) sepia(41%) saturate(7094%) hue-rotate(81deg) brightness(111%) contrast(109%)"
                    crawlDTO["steamids"].push(firstSteamID)
                }

                if (crawlIDs[1] == "") {
                    document.getElementById("isCrawledBeforeCheckMark").style.filter = "invert(78%) sepia(41%) saturate(7094%) hue-rotate(81deg) brightness(111%) contrast(109%)"
                    crawlDTO["steamids"].push(secondSteamID)
                }
                console.log(`crawlDTO is ${JSON.stringify(crawlDTO)}`)
                
                utilRequest.startCrawl(crawlDTO).then(newCrawlIDs => {
                    if (newCrawlIDs.length == 2) {
                        // Two new crawls are starting
                        console.log("both new")
                        console.log(newCrawlIDs)
                        window.location.href = `/crawl?firstcrawlid=${newCrawlIDs[0]}&secondcrawlid=${newCrawlIDs[1]}`
                    } else {
                        crawlIDs = crawlIDs.filter(k => k)
                        console.log(`One or more were existing, going to /crawl?firstcrawlid=${newCrawlIDs[0]}&secondcrawlid=${crawlIDs[0]}`)
                        window.location.href = `/crawl?firstcrawlid=${newCrawlIDs[0]}&secondcrawlid=${crawlIDs[0]}`
                    }
                }, (err) => {
                    console.error(`startCrawl: ${err}`);
                })
            }, errs => {
                console.error(errs)
            })
        }, errs => {
            console.error(errs)
        })
    }
});

function isValidFormatSteamID(steamID) {
    if (steamID.length != 17) {
        return false
    }
    const regex = /([0-9]){17}/g
    const result = steamID.match(regex);
    if (result.length == 1) {
        return true
    }
    return false
}

function displayErrorForInvalidSteamID(isFirstSteamID, errorMessage) {
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

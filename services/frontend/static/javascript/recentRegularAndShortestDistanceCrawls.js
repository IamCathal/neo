import { timezSince } from '/static/javascript/userCard.js';

getAnyNewFinishedCrawlStatuses().then((newCrawls) => {
    let mostRecentCrawls = []
    mostRecentCrawls = newCrawls.concat(mostRecentCrawls)
    mostRecentCrawls.length = mostRecentCrawls.length >= 12 ? 12 : mostRecentCrawls.length;

    renderTopTwelveMostRecentFinishedCrawlStatuses(mostRecentCrawls)
    setInterval(() => {
        getAnyNewFinishedCrawlStatuses().then((res) => {
            mostRecentCrawls = res.concat(mostRecentCrawls)
            mostRecentCrawls.length = mostRecentCrawls.length >= 12 ? 12 : mostRecentCrawls.length;
            renderTopTwelveMostRecentFinishedCrawlStatuses(mostRecentCrawls)
        }, err => {
            console.error(err)
        })
    }, 5000)
}, err => {
    console.error(err)
})

getAnyNewFinishedShortestDistanceCrawlStatuses().then((newCrawls) => {
    let mostRecentShortestDistanceCrawls = []

    mostRecentShortestDistanceCrawls = newCrawls.concat(mostRecentShortestDistanceCrawls)
    mostRecentShortestDistanceCrawls.length = mostRecentShortestDistanceCrawls.length >= 12 ? 12 : mostRecentShortestDistanceCrawls.length;

    renderTopTwelveMostRecentFinishedShortestDistanceCrawlStatuses(mostRecentShortestDistanceCrawls)
    setInterval(() => {
        getAnyNewFinishedCrawlStatuses().then((res) => {
            mostRecentShortestDistanceCrawls = res.concat(mostRecentShortestDistanceCrawls)
            mostRecentShortestDistanceCrawls.length = mostRecentShortestDistanceCrawls.length >= 12 ? 12 : mostRecentShortestDistanceCrawls.length;
            renderTopTwelveMostRecentFinishedCrawlStatuses(mostRecentShortestDistanceCrawls)
        }, err => {
            console.error(err)
        })
    }, 5000)
}, err => {
    console.error(err)
})

getAnyNewFinishedShortestDistanceCrawlStatuses().then((newCrawls) => {
    let mostRecentCrawls = []

    mostRecentCrawls = newCrawls.concat(mostRecentCrawls)
    mostRecentCrawls.length = mostRecentCrawls.length >= 12 ? 12 : mostRecentCrawls.length;

    renderTopTwelveMostRecentFinishedShortestDistanceCrawlStatuses(mostRecentCrawls)
    // setInterval(() => {
    //     getAnyNewFinishedShortestDistanceCrawlStatuses().then((res) => {
    //         renderTopTwelveMostRecentFinishedShortestDistanceCrawlStatuses(res)
    //     }, err => {
    //         console.error(err)
    //     })
    // }, 5000)
}, err => {
    console.error(err)
})

function renderTopTwelveMostRecentFinishedCrawlStatuses(mostRecentCrawls) {
    document.getElementById("finishedCrawlsDiv").innerHTML = '';
    const backgroundColors = ['#292929', '#414141']
    let i = 0;
    mostRecentCrawls.forEach(crawl => {
        const usersFlagEmoji = getFlagEmoji(crawl.user.accdetails.loccountrycode) == "" ? '🏴‍☠️' : getFlagEmoji(crawl.user.accdetails.loccountrycode)
        const creationDate = new Date(crawl.crawlingstatus.timestarted*1000);
        const timeSinceString = `${timezSince(creationDate)}`

        document.getElementById("finishedCrawlsDiv").innerHTML += `
        <div class="row pt-1 pb-1 mt-1" style="border-radius: 8px; background-color:${backgroundColors[i%backgroundColors.length]};">
                        <div class="col-sm- ml-2">
                            <img src="${crawl.user.accdetails.avatar}">
                        </div>
                        <div class="col-sm- ml-2">
                            ${usersFlagEmoji}
                        </div>
                        <div class="col-5">
                            ${crawl.user.accdetails.personaname}
                        </div>
                        <div class="col-2">
                            ${crawl.crawlingstatus.userscrawled} users
                        </div>
                        <div class="col-2">
                            ${timeSinceString}
                        </div>
                        <div class="col-sm- ml-3">
                            <a href="/graph/${crawl.crawlingstatus.crawlid}">View now
                        </div>
                    </div>
        `
        i++;
    })
}

function renderTopTwelveMostRecentFinishedShortestDistanceCrawlStatuses(mostRecentCrawls) {
    document.getElementById("finishedShortestDistanceCrawlsDiv").innerHTML = '';
    mostRecentCrawls.forEach(crawl => {
        const firstUsersFlagEmoji = getFlagEmoji(crawl.firstuser.accdetails.loccountrycode) == "" ? '🏴‍☠️' : getFlagEmoji(crawl.firstuser.accdetails.loccountrycode)
        const secondUsersFlagEmoji = getFlagEmoji(crawl.seconduser.accdetails.loccountrycode) == "" ? '🏴‍☠️' : getFlagEmoji(crawl.seconduser.accdetails.loccountrycode)
        const firstUserMediumQualityProfiler = crawl.firstuser.accdetails.avatar.split(".jpg").join("") + "_medium.jpg";
        const secondUserMediumQualityProfiler = crawl.seconduser.accdetails.avatar.split(".jpg").join("") + "_medium.jpg";

        const crawlStartTime = new Date(crawl.timestarted*1000);
        const timeSinceString = `${timezSince(crawlStartTime)}`

        document.getElementById("finishedShortestDistanceCrawlsDiv").innerHTML += `
        <div class="col-5" style="border: 2px solid white; border-radius: 8px;">
                            <div class="row pt-2">
                                <div class="col text-center">
                                    <img 
                                        src="${firstUserMediumQualityProfiler}"
                                    >
                                </div>
                                <div class="col text-center">
                                    <img 
                                        src="${secondUserMediumQualityProfiler}"
                                    >
                                </div>
                            </div>
                            <div class="row">
                                <div class="col text-center">
                                    ${firstUsersFlagEmoji}  ${crawl.firstuser.accdetails.personaname} 
                                </div>
                                <div class="col text-center">
                                    ${secondUsersFlagEmoji}  ${crawl.seconduser.accdetails.personaname} 
                                </div>
                            </div>
                            <div class="row">
                                <div class="col text-center">
                                    <p style="font-size: 1.3rem;">${crawl.totalnetworkspan} users</p>
                                </div>
                                <div class="col text-center">
                                    <p style="font-size: 1.3rem;">${timeSinceString} ago</p>
                                </div>
                            </div>
                            <div class="row justify-content-md-center mb-2">
                               <div class="col-3 text-center ml-2 mr-2 mt-1 mb-1" style="background-color: aqua; border-radius: 6px;">
                                    View
                               </div>
                            </div>
                        </div>
        `
    })
}


function getAnyNewFinishedCrawlStatuses() {
    return new Promise((resolve, reject) => {
        fetch(`http://localhost:2590/api/getfinishedcrawlsaftertimestamp?timestamp=${5}`, {
            headers: {
                "Content-Type": "application/json"
            },
        }).then((res => res.json()))
        .then(data => {
            resolve(data.crawls)
        }).catch(err => {
            reject(err)
        })
    })
}

function getAnyNewFinishedShortestDistanceCrawlStatuses() {
    return new Promise((resolve, reject) => {
        fetch(`http://localhost:2590/api/getfinishedshortestdistancecrawlsaftertimestamp?timestamp=${-5}`, {
            headers: {
                "Content-Type": "application/json"
            },
        }).then((res => res.json()))
        .then(data => {
            resolve(data.crawlingstatus)
        }).catch(err => {
            reject(err)
        })
    })
}

// COMMON
// https://dev.to/jorik/country-code-to-flag-emoji-a21
function getFlagEmoji(countryCode) {
    const codePoints = countryCode
      .toUpperCase()
      .split('')
      .map(char =>  127397 + char.charCodeAt());
    return String.fromCodePoint(...codePoints);
}
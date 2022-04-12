import * as util from '/static/javascript/util.js';
import * as utilRequest from '/static/javascript/utilRequests.js';

utilRequest.getAnyNewFinishedCrawlStatuses().then((newCrawls) => {
    let mostRecentCrawls = []
    mostRecentCrawls = newCrawls.concat(mostRecentCrawls)
    mostRecentCrawls.length = mostRecentCrawls.length >= 12 ? 12 : mostRecentCrawls.length;

    renderTopTwelveMostRecentFinishedCrawlStatuses(mostRecentCrawls)
    setInterval(() => {
        utilRequest.getAnyNewFinishedCrawlStatuses().then((res) => {
            mostRecentCrawls = res.concat(mostRecentCrawls)
            mostRecentCrawls.length = mostRecentCrawls.length >= 12 ? 12 : mostRecentCrawls.length;
            renderTopTwelveMostRecentFinishedCrawlStatuses(mostRecentCrawls)
        }, err => {
            console.error(err)
        })
    }, 30000)
}, err => {
    console.error(err)
})

utilRequest.getAnyNewFinishedShortestDistanceCrawlStatuses().then((newCrawls) => {
    let mostRecentShortestDistanceCrawls = []

    mostRecentShortestDistanceCrawls = newCrawls.concat(mostRecentShortestDistanceCrawls)
    mostRecentShortestDistanceCrawls.length = mostRecentShortestDistanceCrawls.length >= 12 ? 12 : mostRecentShortestDistanceCrawls.length;

    renderTopTwelveMostRecentFinishedShortestDistanceCrawlStatuses(mostRecentShortestDistanceCrawls)
    setInterval(() => {
        utilRequest.getAnyNewFinishedCrawlStatuses().then((res) => {
            mostRecentShortestDistanceCrawls = res.concat(mostRecentShortestDistanceCrawls)
            mostRecentShortestDistanceCrawls.length = mostRecentShortestDistanceCrawls.length >= 12 ? 12 : mostRecentShortestDistanceCrawls.length;
            renderTopTwelveMostRecentFinishedCrawlStatuses(mostRecentShortestDistanceCrawls)
        }, err => {
            console.error(err)
        })
    }, 30000)
}, err => {
    console.error(err)
})

utilRequest.getAnyNewFinishedShortestDistanceCrawlStatuses().then((newCrawls) => {
    let mostRecentCrawls = []

    mostRecentCrawls = newCrawls.concat(mostRecentCrawls)
    mostRecentCrawls.length = mostRecentCrawls.length >= 12 ? 12 : mostRecentCrawls.length;

    renderTopTwelveMostRecentFinishedShortestDistanceCrawlStatuses(mostRecentCrawls)
    // setInterval(() => {
    //     utilRequest.getAnyNewFinishedShortestDistanceCrawlStatuses().then((res) => {
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
        const usersFlagEmoji = util.getFlagEmoji(crawl.user.accdetails.loccountrycode) == "" ? '🏴‍☠️' : getFlagEmoji(crawl.user.accdetails.loccountrycode)
        const creationDate = new Date(crawl.crawlingstatus.timestarted*1000);
        const timeSinceString = `${util.timeSinceShort(creationDate)}`

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
            <div class="col-sm- mr-2">
                ${crawl.crawlingstatus.maxlevel}
            </div>
            <div class="col-2 pl-5">
                ${timeSinceString} ago
            </div>
            <div class="col-sm- ml-5 align-self-end">
                <div class="col-3 pr-5" style="border: 2px solid white; border-radius: 6px;">
                    <a href="/graph/${crawl.crawlingstatus.crawlid}" style="text-decoration: none">
                        View
                    </a>
                </div>
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
        const timeSinceString = `${util.timeSinceShort(crawlStartTime)}`

        document.getElementById("finishedShortestDistanceCrawlsDiv").innerHTML += `
        <div class="col-5 m-3 box" style="border: 2px solid white; border-radius: 8px;">
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
                    <a href="/shortestdistance?firstcrawlid=${crawl.crawlids[0]}&secondcrawlid=${crawl.crawlids[1]}" style="text-decoration: none"> View </a>
                </div>
            </div>
        </div>
        `
    })
}
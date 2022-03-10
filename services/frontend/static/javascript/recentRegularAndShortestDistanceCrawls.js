import { timezSince } from '/static/javascript/userCard.js';

getAnyNewFinishedCrawlStatuses().then((newCrawls) => {
    let mostRecentCrawls = []
    console.log(`new crawls ${newCrawls} ${newCrawls.length}`)

    mostRecentCrawls = newCrawls.concat(mostRecentCrawls)
    mostRecentCrawls.length = mostRecentCrawls.length >= 12 ? 12 : mostRecentCrawls.length;
    console.log(`most recent crawls: ${mostRecentCrawls} ${mostRecentCrawls.length}`)
    console.log(mostRecentCrawls)

    renderTopTwelveMostRecentFinishedCrawlStatuses(mostRecentCrawls)
    // setInterval(() => {
    //     getAnyNewFinishedCrawlStatuses().then((res) => {
    //         renderTopTwelveMostRecentFinishedCrawlStatuses(res)
    //     }, err => {
    //         console.error(err)
    //     })
    // }, 5000)
}, err => {
    console.error(err)
})


function renderTopTwelveMostRecentFinishedCrawlStatuses(mostRecentCrawls) {
    const backgroundColors = ['#292929', '#414141']
    let i = 0;
    mostRecentCrawls.forEach(crawl => {
        const usersFlagEmoji = getFlagEmoji(crawl.user.accdetails.loccountrycode) == "" ? '🏴‍☠️' : getFlagEmoji(crawl.user.accdetails.loccountrycode)
        const creationDate = new Date(crawl.crawlingstatus.timestarted*1000);
        const dateString = `${creationDate.getDate()} ${creationDate.toLocaleString('default', { month: 'long' })} ${creationDate.getFullYear()}`;
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


function getAnyNewFinishedCrawlStatuses() {
    return new Promise((resolve, reject) => {
        fetch(`http://localhost:2590/api/getfinishedcrawlsaftertimestamp?timestamp=${5}`, {
            headers: {
                "Content-Type": "application/json"
            },
        }).then((res => res.json()))
        .then(data => {
            console.log(data.crawls)
            resolve(data.crawls)
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
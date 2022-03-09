let twelveMostRecentFinishedCrawls = []

getAnyNewFinishedCrawlStatuses(response => {
    console.log(response)
    setInterval(() => {
        getAnyNewFinishedCrawlStatuses(response => {
            console.log(response)
        }, err => {
            console.error(err)
        })
    }, 300000)
}, err => {
    console.error(err)
})

function fillInRecentFinishedCrawlStatuses(newCrawlStatuses) {
    twelveMostRecentFinishedCrawls = twelveMostRecentFinishedCrawls.unshift(newCrawlStatuses)
    twelveMostRecentFinishedCrawls = twelveMostRecentFinishedCrawls(12)
}


function getAnyNewFinishedCrawlStatuses() {
    return new Promise((resolve, reject) => {
        fetch(`http://localhost:2590/api/getfinishedcrawlsaftertimestamp?timestamp=${5}`, {
            headers: {
                "Content-Type": "application/json"
            },
        }).then((res => res.json()))
        .then(data => {
            resolve(data)
        }).catch(err => {
            console.error(err);
        })
    })
}
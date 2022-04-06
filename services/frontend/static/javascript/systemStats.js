import { countUpElement } from '/static/javascript/countUpScript.js';


getTotalUsersInDB().then(totalUsers => {
    let totalUsersCached = totalUsers
    countUpElement("totalUsersCrawled", totalUsersCached)
    setInterval(() => {
        getTotalUsersInDB().then((newTotalUsers) => {
            if (newTotalUsers != totalUsersCached) {
                totalUsersCached = newTotalUsers
                const currentVal = parseInt(document.getElementById("totalUsersCrawled").textContent.replaceAll(",",""))
                countUpElement("totalUsersCrawled", totalUsersCached, {'startVal': currentVal})
            }
        }, err => {
            console.error(err)
        })
    }, 30000)
})

getTotalCrawls().then(totalCrawls => {
    let totalCrawlsCached = totalCrawls
    countUpElement("totalCrawls", totalCrawlsCached)
    setInterval(() => {
        getTotalCrawls().then((newTotalCrawls) => {
            if (newTotalCrawls != totalCrawlsCached) {
                totalCrawlsCached = newTotalCrawls
                const currentVal = parseInt(document.getElementById("totalCrawls").textContent.replaceAll(",",""));
                countUpElement("totalCrawls", totalCrawlsCached, {'startVal': currentVal})
            }
        }, err => {
            console.error(err)
        })
    }, 30000)
}, err => {
    console.error(err)
})

function getTotalUsersInDB() {
    return new Promise((resolve, reject) => {
        fetch(`http://localhost:2590/api/gettotalusersindb`, {
            headers: {
                "Content-Type": "application/json"
            },
        }).then((res => res.json()))
        .then(data => {
            resolve(data.usersindb)
        }).catch(err => {
            reject(err)
        })
    })
}

function getTotalCrawls() {
    return new Promise((resolve, reject) => {
        fetch(`http://localhost:2590/api/gettotalcrawlscompleted`, {
            headers: {
                "Content-Type": "application/json"
            },
        }).then((res => res.json()))
        .then(data => {
            resolve(data.totalcrawls)
        }).catch(err => {
            reject(err)
        })
    })
}
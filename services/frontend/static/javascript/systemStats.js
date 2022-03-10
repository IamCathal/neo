import { countUpElement } from '/static/javascript/countUpScript.js';


getTotalUsersInDB().then(totalUsers => {
    countUpElement("totalUsersCrawled", totalUsers)
})

getTotalCrawls().then(totalCrawls => {
    countUpElement("totalCrawls", totalCrawls)
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
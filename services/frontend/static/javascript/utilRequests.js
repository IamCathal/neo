export function doesProcessedGraphDataExist(crawlID) {
    return new Promise((resolve, reject) => {
        fetch(`http://localhost:2590/api/doesprocessedgraphdataexist/${crawlID}`, {
            method: 'POST',
            headers: {
                "Content-Type": "application/json"
            },
        }).then((res => res.json()))
        .then(data => {
            if (data.exists == "yes") {
                resolve(true)
            } 
            resolve(false)
        }).catch(err => {
            reject(err)
        })
    });
}

export function getProcessedGraphData(crawlID) {
    return new Promise((resolve, reject) => {
        fetch(`http://localhost:2590/api/getprocessedgraphdata/${crawlID}`, {
            method: "POST",
            headers: {
                'Accept-Encoding': 'gzip'
            }
        }).then(res => res.json())
        .then(data => {
            resolve(data)
        }).catch(err => {
            reject(err)
        })
    });
}

export function getAnyNewFinishedCrawlStatuses() {
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

export function getAnyNewFinishedShortestDistanceCrawlStatuses() {
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

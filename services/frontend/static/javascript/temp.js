function getCrawlingUserWhenAvailable() {
    return new Promise((resolve, reject) => {
    let interval = setInterval(() => {
        if (usersCrawledIsMoreThanZero()) {
            clearInterval(interval);
            getUser(crawlID).then(user => {
                setUserCardDetails(user)
                resolve(true)
            }, err => {
                reject(err)
            })
        }
      }, 50);
    });
}

function usersCrawledIsMoreThanZero() {
    if (parseInt(document.getElementById("usersCrawled").textContent) >= 1) {
        return true
    }
    return false
}

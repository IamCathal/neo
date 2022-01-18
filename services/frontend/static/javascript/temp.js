function getCrawlingUserWhenAvailable() {
    return new Promise((resolve, reject) => {
    
    let interval = setInterval(() => {
        console.log("stuck in a loop man")
        if (usersCrawledIsMoreThanZero) {
            console.log("out of the loop man")
            clearInterval(interval);
            getUser(crawlID).then(user => {
                console.log("got user!")
                console.log(user)
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
    if (!isNaN((parseInt(document.getElementById("usersCrawled").textContent))) &&
        parseInt(document.getElementById("usersCrawled").textContent) >= 2) {
            return true
        }
    return false
}

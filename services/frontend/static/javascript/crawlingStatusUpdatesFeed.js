function initAndMonitorCrawlingStatusWebsocket(crawlID) {

    let wsConn = new WebSocket(`ws://localhost:2590/ws/crawlingstatstream/${crawlID}`);
    wsConn.addEventListener("close", (evt) => {
        console.log("CLOSED!");
    })
    console.log(wsConn)
    wsConn.addEventListener("message", (evt) => {
        let crawlingStatUpdate = JSON.parse(evt.data);
        document.getElementById("usersCrawled").textContent = crawlingStatUpdate.userscrawled;
        document.getElementById("totalUsersToCrawl").textContent = crawlingStatUpdate.totaluserstocrawl;
        document.getElementById("percentageDone").textContent = `${Math.floor((crawlingStatUpdate.userscrawled/crawlingStatUpdate.totaluserstocrawl)*100)}%`;
        document.getElementById("crawlTime").textContent = timeSince(new Date(crawlingStatUpdate.timestarted*1000));
        
        document.getElementById("progressBarID").style.width = `${Math.floor((crawlingStatUpdate.userscrawled/crawlingStatUpdate.totaluserstocrawl)*100)}%`;
    })
}

function timeSince(targetDate) {
    let seconds = Math.floor((new Date()-targetDate)/1000)
    let interval = seconds / 31536000 // years
    interval = seconds / 2592000; // months
    interval = seconds / 86400; // days
    if (interval > 1) {
      return Math.floor(interval) + "d ago";
    }
    interval = seconds / 3600;
    if (interval > 1) {
      return Math.floor(interval) + "h ago";
    }
    interval = seconds / 60;
    if (interval > 1) {
      return Math.floor(interval) + "m ago";
    }
    return Math.floor(seconds) + "s";
}
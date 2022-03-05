export function initAndMonitorCrawlingStatusWebsocket(crawlID, idPrefix, isAlreadyDone) {
  return new Promise((resolve, reject) => {
    if (isAlreadyDone) {
      resolve()
    }

    let wsConn = new WebSocket(`ws://localhost:2590/ws/crawlingstatstream/${crawlID}`);
    wsConn.addEventListener("close", (evt) => {
        console.log("CLOSED!");
    })
    
    wsConn.addEventListener("message", (evt) => {
        let crawlingStatUpdate = JSON.parse(evt.data);
        if (crawlingStatUpdate.userscrawled === crawlingStatUpdate.totaluserstocrawl) {
            document.getElementById(`${idPrefix}CrawlStatus`).textContent = "Processing graph"
            wsConn.close()
            resolve()
        }

        document.getElementById(`${idPrefix}CrawlStatus`).textContent = 'Crawling'
        document.getElementById(`${idPrefix}UsersCrawled`).textContent = crawlingStatUpdate.userscrawled;
        document.getElementById(`${idPrefix}TotalUsersToCrawl`).textContent = crawlingStatUpdate.totaluserstocrawl;
        document.getElementById(`${idPrefix}PercentageDone`).textContent = `${Math.floor((crawlingStatUpdate.userscrawled/crawlingStatUpdate.totaluserstocrawl)*100)}%`;
        document.getElementById(`${idPrefix}CrawlTime`).textContent = timeSince(new Date(crawlingStatUpdate.timestarted*1000));
        document.getElementById(`${idPrefix}ProgressBarID`).style.width = `${Math.floor((crawlingStatUpdate.userscrawled/crawlingStatUpdate.totaluserstocrawl)*100)}%`;
    
        document.title = `${crawlingStatUpdate.userscrawled}/${crawlingStatUpdate.totaluserstocrawl} - Crawling`;
      })
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
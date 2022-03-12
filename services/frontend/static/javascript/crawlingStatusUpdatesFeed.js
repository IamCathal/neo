import * as utilRequest from '/static/javascript/utilRequests.js';

export function initAndMonitorCrawlingStatusWebsocket(crawlID, idPrefix, isAlreadyDone) {
  return new Promise((resolve, reject) => {
    if (isAlreadyDone) {
      utilRequest.getCrawlingStatus(crawlID).then(crawlingStatus => {

        document.getElementById(`${idPrefix}CrawlStatus`).textContent = 'Completed'
        document.getElementById(`${idPrefix}UsersCrawled`).textContent = crawlingStatus.userscrawled;
        document.getElementById(`${idPrefix}TotalUsersToCrawl`).textContent = crawlingStatus.totaluserstocrawl;
        document.getElementById(`${idPrefix}PercentageDone`).textContent = `${Math.floor((crawlingStatus.userscrawled/crawlingStatus.totaluserstocrawl)*100)}%`;
        document.getElementById(`${idPrefix}CrawlTime`).textContent = timeSince(new Date(crawlingStatus.timestarted*1000));
        document.getElementById(`${idPrefix}ProgressBarID`).style.width = `${Math.floor((crawlingStatus.userscrawled/crawlingStatus.totaluserstocrawl)*100)}%`;
        
        resolve()
      }, err => {
        reject(err)
      })
    }

    let wsConn = new WebSocket(`ws://localhost:2590/ws/crawlingstatstream/${crawlID}`);
    wsConn.addEventListener("close", (evt) => {
        console.log("CLOSED!");
    })
    
    wsConn.addEventListener("message", (evt) => {
        const crawlingStatUpdate = JSON.parse(evt.data);
        if (crawlingStatUpdate.userscrawled == crawlingStatUpdate.totaluserstocrawl) {
            setCrawlingStatusToProcessing(idPrefix).then((res) => {
              wsConn.close()
              resolve()
            })
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

function setCrawlingStatusToProcessing(idPrefix) {
  return new Promise((resolve, reject) =>{
    document.getElementById(`${idPrefix}CrawlStatus`).textContent = 'Processing graph';
    resolve(true);
  });
}
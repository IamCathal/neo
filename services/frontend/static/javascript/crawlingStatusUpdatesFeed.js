import { startCreateGraph, doesProcessedGraphDataExist } from '/static/javascript/crawlingPage.js'

export function initAndMonitorCrawlingStatusWebsocket(crawlID) {

    let wsConn = new WebSocket(`ws://localhost:2590/ws/crawlingstatstream/${crawlID}`);
    wsConn.addEventListener("close", (evt) => {
        console.log("CLOSED!");
    })

    wsConn.addEventListener("message", (evt) => {
        let crawlingStatUpdate = JSON.parse(evt.data);
        if (crawlingStatUpdate.userscrawled === crawlingStatUpdate.totaluserstocrawl) {
            document.getElementById("crawlStatus").textContent = "Processing graph"

            startCreateGraph(crawlingStatUpdate.crawlid).then(res => {
              console.log(`create graph: ${res}`);
              // graph is now being created. Wait for completion then redirect
              
              // Check every 400ms is the graph is done processing yet
              let interval = setInterval(function() {
                doesProcessedGraphDataExist(crawlingStatUpdate.crawlid).then(doesExist => {
                  if (doesExist === true) {
                    clearInterval(interval);
                    window.location.href = `/graph/${crawlingStatUpdate.crawlid}`;
                  } else {
                    console.log("graph not done processing")
                  }
                }, err => {
                  clearInterval(interval);
                  console.error(`error checking if graph data is procced: ${err}`);
                })
              }, 500);

              
            }, err => {
              console.error(`err from createGraph ${err}`)
            });
        }

        
        document.getElementById("crawlStatus").textContent = 'Crawling'
        
        // console.log(document.getElementById("crawlStatus"))
        // if (document.getElementById("crawlStatus") == 'Initialising') {
        //   console.log("is init")
        // } else {
        //   console.log("isn't init")
        // }
        document.getElementById("usersCrawled").textContent = crawlingStatUpdate.userscrawled;
        document.getElementById("totalUsersToCrawl").textContent = crawlingStatUpdate.totaluserstocrawl;
        document.getElementById("percentageDone").textContent = `${Math.floor((crawlingStatUpdate.userscrawled/crawlingStatUpdate.totaluserstocrawl)*100)}%`;
        document.getElementById("crawlTime").textContent = timeSince(new Date(crawlingStatUpdate.timestarted*1000));
        document.getElementById("progressBarID").style.width = `${Math.floor((crawlingStatUpdate.userscrawled/crawlingStatUpdate.totaluserstocrawl)*100)}%`;
    
        document.title = `${crawlingStatUpdate.userscrawled}/${crawlingStatUpdate.totaluserstocrawl} - Crawling`;
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
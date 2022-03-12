import { initAndMonitorCrawlingStatusWebsocket } from '/static/javascript/crawlingStatusUpdatesFeed.js'
import * as util from '/static/javascript/util.js';
import * as utilRequest from '/static/javascript/utilRequests.js';

const crawlIDs = util.getCrawlIDs()

if (crawlIDs.length == 2) {
    utilRequest.getShortestDistanceInfo(crawlIDs).then(shortestDistanceInfo => {
        if (shortestDistanceInfo.crawlids === null) {
            // Shortest distance has not been calculated yet
            renderCrawlStatusBoxes(2)

            // Get crawling status of both, either none are finished or one has already been crawled before
            const firstUserIsDone = utilRequest.isCrawlingFinished(crawlIDs[0])
            const secondUserIsDone = utilRequest.isCrawlingFinished(crawlIDs[1])

            Promise.all([firstUserIsDone, secondUserIsDone]).then(isDones => {

                let initAndMonitorCrawlOne = initAndMonitorCrawlingStatusWebsocket(crawlIDs[0], "firstCrawl", isDones[0])
                let initAndMonitorCrawlTwo = initAndMonitorCrawlingStatusWebsocket(crawlIDs[1], "secondCrawl", isDones[1])
    
                utilRequest.getCrawlingUserWhenAvailable(crawlIDs[0], "firstCrawl").then(res => {}, err => {
                    console.error(`error getting first crawling user: ${JSON.stringify(err)}`)
                })
                utilRequest.getCrawlingUserWhenAvailable(crawlIDs[1], "secondCrawl").then(res => {}, err => {
                    console.error(`error getting second crawling user: ${JSON.stringify(err)}`)
                })
    
                Promise.all([initAndMonitorCrawlOne, initAndMonitorCrawlTwo]).then(vals => {
                    let firstUserStartCreateGraph = utilRequest.startCreateGraph(crawlIDs[0])
                    let secondUserStartCreateGraph = utilRequest.startCreateGraph(crawlIDs[1])
    
                    Promise.all([firstUserStartCreateGraph, secondUserStartCreateGraph]).then(vals => {
                        let firstUserGraphExists = utilRequest.waitUntilGraphDataExists(crawlIDs[0])
                        let secondUserGraphExists = utilRequest.waitUntilGraphDataExists(crawlIDs[1])
    
                        Promise.all([firstUserGraphExists, secondUserGraphExists]).then(vals => {
                            utilRequest.startCalculateGetShortestDistance(crawlIDs).then(res => {
                                console.log(res);
                                window.location.href = `/shortestdistance?firstcrawlid=${crawlIDs[0]}&secondcrawlid=${crawlIDs[1]}`
                            }, err => {
                                console.error(`err calculating shortest distance: ${JSON.stringify(err)}`)
                            })
                            
                        }, err => {
                            console.error(`error waiting until both graphs existed: ${JSON.stringify(err)}`)
                        })
                    }, errs => {
                        console.error(errs)
                    })
                }, err => {
                    console.error(err)
                })
            }, (errs) => {
                console.error(`error(s) calling utilRequest.isCrawlingFinished: ${errs}`)
            })
        } else {
            window.location.href = `/shortestdistance?firstcrawlid=${crawlIDs[0]}&secondcrawlid=${crawlIDs[1]}`
        }
    }, err => {
        console.error(err)
    })
}

if (crawlIDs.length == 1) {
    utilRequest.doesProcessedGraphDataExist(crawlIDs[0]).then(doesExist => {
        if (doesExist) {
            console.log("did exist")
            // forward to that page
            // window.location.href = `/graph/${crawlID}`;
        } else {
            console.log("no exist")
            renderCrawlStatusBoxes(1)
            // subscribe to crawling status updates
            initAndMonitorCrawlingStatusWebsocket(crawlIDs[0], "firstCrawl").then(res => {
                utilRequest.startCreateGraph(crawlIDs[0]).then(res => {
              
                // Check every 500ms is the graph is done processing yet
                let interval = setInterval(function() {
                    utilRequest.doesProcessedGraphDataExist(crawlIDs[0]).then(doesExist => {
                        if (doesExist === true) {
                            clearInterval(interval);
                            window.location.href = `/graph/${crawlIDs[0]}`;
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
            }, err => {
                console.error(err)
            })
            utilRequest.getCrawlingUserWhenAvailable(crawlIDs[0], "firstCrawl").then(res => {}, err => {
                console.error(`error getting crawling user: ${err}`)
            })
        }
    }, err => {
        console.error(`error calling get finished graph data: ${err}`);
    })
} 

function renderCrawlStatusBoxes(numberOfBoxes) {
    if (numberOfBoxes == 1) {
        // Add spacing to left
        document.getElementById("crawlBoxRow").innerHTML += `<!-- Left Spacing -->
        <div class="col-3"></div>`;
        // main crawl status box
        document.getElementById("crawlBoxRow").innerHTML += `
        <div class="col box pt-3 pl-4 pr-4">

                <div class="row mt-2">
                    <div class="col">
                        <div class="row text-center">
                            <p> Users Crawled </p>
                        </div>
                        <div class="row text-center mt-1">
                            <p style="font-size: 4.3rem; font-weight: 600;" id="firstCrawlUsersCrawled"> 0 </p>
                        </div>
                    </div>
                    <div class="col">
                        <div class="row text-center">
                            <p>  Total Users To Crawl </p>
                        </div>
                        <div class="row text-center">
                            <p style="font-size: 4.3rem; font-weight: 600;" id="firstCrawlTotalUsersToCrawl" > 0 </p>
                        </div>
                    </div>
                </div>

                <div class="row mt-2">
                    <div class="col">
                        <div class="row text-center">
                            <p> Status </p>
                        </div>
                        <div class="row text-center">
                            <p style="font-size: 1.9rem; font-weight: 600;" id="firstCrawlCrawlStatus"> Initialising </p>
                        </div>
                    </div>
                    <div class="col">
                        <div class="row text-center">
                            <p>  Started Crawl </p>
                        </div>
                        <div class="row text-center">
                            <p style="font-size: 1.9rem; font-weight: 600;" id="firstCrawlCrawlTime"> 0s </p>
                        </div>
                    </div>
                </div>

                <!-- crawling progress bar -->
                <div class="row crawlProgressBar mt-4">
                    <span class="progressBarSpan" id="firstCrawlProgressBarID"></span>
                </div>

                <!-- crawling progress bar -->
                <div class="row text-center mt-1">
                    <div class="col" style="font-size: 0.9rem;" id="firstCrawlPercentageDone"> 0% </div>
                </div>
                    


                <hr class="mt-3 mb-4">

                <!-- User info card -->
                <div class="row mt-3 mb-3">
                    <div class="col box">
                        <div class="row ml-0">
                            <div class="col-8">
                                <div class="row">
                                    <div class="col mb-2">
                                        <p id="firstCrawlUserUsername" class="skeleton" style="font-size: 2rem; font-weight: 700; width: 80%; height: 3rem" >
                                            
                                        </p>
                                    </div>
                                </div>
                                <div class="row">
                                    <div class="col-4">
                                        <p>
                                            Country:
                                        </p>
                                    </div>
                                    <div class="col">
                                        <p id="firstCrawlUserCountry" style="font-weight: 500" class="skeleton skeleton-text truncate"></p>
                                            
                                        </p>
                                    </div>
                                </div>
                                <div class="row">
                                    <div class="col-4">
                                        <p>
                                            Friends: 
                                        </p>
                                    </div>
                                    <div class="col">
                                        <p id="firstCrawlUserFriendCount" style="font-weight: 500; width: 90%" class="skeleton skeleton-text">
                                            
                                        </p>
                                    </div>
                                </div>
                                <div class="row">
                                    <div class="col-4">
                                        <p>
                                            Created:
                                        </p>
                                    </div>
                                    <div class="col">
                                        <p id="firstCrawlUserCreationDate" style="font-weight: 500" class="skeleton skeleton-text">
                                            
                                        </p>
                                    </div>
                                </div>
                                <div class="row">
                                    <div class="col-4">
                                        <p>
                                            Profile: 
                                        </p>
                                    </div>
                                    <div class="col">
                                        <p id="firstCrawlUserProfile" style="font-weight: 500; width: 83%" class="skeleton skeleton-text">
                                            
                                        </p>
                                    </div>
                                </div>
                            </div>
                            
                            <div class="col" id="firstCrawlUserAvatarCol">
                                <div class="firstCrawlUserProfiler" style="position: relative">
                                    <img class="mr-1 float-end skeleton" style="border-radius: 10px; height: 10rem; width: 10rem" id="firstCrawlUserAvatar"></img>
                                </div>
                                <div class="firstCrawlUserCountry " style="position: absolute; bottom: 0; left: 12%;">
                                    <img id="firstCrawlUserCountryFlag" class="float-end" style="width: 50%"></img>
                                </div>
                            </div>
                        </div>
                    </div>
                </div> <!-- end user info card -->
            </div>
        `;
        // Add spacing to right
        document.getElementById("crawlBoxRow").innerHTML += `<!-- Right Spacing -->
        <div class="col-3"></div>`;
        return
    }

    // first user crawl status box
    document.getElementById("crawlBoxRow").innerHTML += `
    <div class="col box pt-3 pl-4 pr-4">

            <div class="row mt-2">
                <div class="col">
                    <div class="row text-center">
                        <p> Users Crawled </p>
                    </div>
                    <div class="row text-center mt-1">
                        <p style="font-size: 4.3rem; font-weight: 600;" id="firstCrawlUsersCrawled"> 0 </p>
                    </div>
                </div>
                <div class="col">
                    <div class="row text-center">
                        <p>  Total Users To Crawl </p>
                    </div>
                    <div class="row text-center">
                        <p style="font-size: 4.3rem; font-weight: 600;" id="firstCrawlTotalUsersToCrawl" > 0 </p>
                    </div>
                </div>
            </div>

            <div class="row mt-2">
                <div class="col">
                    <div class="row text-center">
                        <p> Status </p>
                    </div>
                    <div class="row text-center">
                        <p style="font-size: 1.9rem; font-weight: 600;" id="firstCrawlCrawlStatus"> Initialising </p>
                    </div>
                </div>
                <div class="col">
                    <div class="row text-center">
                        <p>  Started Crawl </p>
                    </div>
                    <div class="row text-center">
                        <p style="font-size: 1.9rem; font-weight: 600;" id="firstCrawlCrawlTime"> 0s </p>
                    </div>
                </div>
            </div>

            <!-- crawling progress bar -->
            <div class="row crawlProgressBar mt-4">
                <span class="progressBarSpan" id="firstCrawlProgressBarID"></span>
            </div>

            <!-- crawling progress bar -->
            <div class="row text-center mt-1">
                <div class="col" style="font-size: 0.9rem;" id="firstCrawlPercentageDone"> 0% </div>
            </div>
                


            <hr class="mt-3 mb-4">

            <!-- User info card -->
            <div class="row mt-3 mb-3">
                <div class="col box">
                    <div class="row ml-0">
                        <div class="col-8">
                            <div class="row">
                                <div class="col mb-2">
                                    <p id="firstCrawlUserUsername" class="skeleton" style="font-size: 2rem; font-weight: 700; width: 80%; height: 3rem" >
                                        
                                    </p>
                                </div>
                            </div>
                            <div class="row">
                                <div class="col-4">
                                    <p>
                                        Country:
                                    </p>
                                </div>
                                <div class="col">
                                    <p id="firstCrawlUserCountry" style="font-weight: 500" class="skeleton skeleton-text truncate"></p>
                                        
                                    </p>
                                </div>
                            </div>
                            <div class="row">
                                <div class="col-4">
                                    <p>
                                        Friends: 
                                    </p>
                                </div>
                                <div class="col">
                                    <p id="firstCrawlUserFriendCount" style="font-weight: 500; width: 90%" class="skeleton skeleton-text">
                                        
                                    </p>
                                </div>
                            </div>
                            <div class="row">
                                <div class="col-4">
                                    <p>
                                        Created:
                                    </p>
                                </div>
                                <div class="col">
                                    <p id="firstCrawlUserCreationDate" style="font-weight: 500" class="skeleton skeleton-text">
                                        
                                    </p>
                                </div>
                            </div>
                            <div class="row">
                                <div class="col-4">
                                    <p>
                                        Profile: 
                                    </p>
                                </div>
                                <div class="col">
                                    <p id="firstCrawlUserProfile" style="font-weight: 500; width: 83%" class="skeleton skeleton-text">
                                        
                                    </p>
                                </div>
                            </div>
                        </div>
                        
                        <div class="col" id="firstCrawlUserAvatarCol">
                            <div class="firstCrawlUserProfiler" style="position: relative">
                                <img class="mr-1 float-end skeleton" style="border-radius: 10px; height: 10rem; width: 10rem" id="firstCrawlUserAvatar"></img>
                            </div>
                            <div class="firstCrawlUserCountry " style="position: absolute; bottom: 0; left: 12%;">
                                <img id="firstCrawlUserCountryFlag" class="float-end" style="width: 50%"></img>
                            </div>
                        </div>
                    </div>
                </div>
            </div> <!-- end user info card -->
        </div>
    `;
    // Second user status crawl box
    document.getElementById("crawlBoxRow").innerHTML += `
    <div class="col box pt-3 pl-4 pr-4 ml-3">

            <div class="row mt-2">
                <div class="col">
                    <div class="row text-center">
                        <p> Users Crawled </p>
                    </div>
                    <div class="row text-center mt-1">
                        <p style="font-size: 4.3rem; font-weight: 600;" id="secondCrawlUsersCrawled"> 0 </p>
                    </div>
                </div>
                <div class="col">
                    <div class="row text-center">
                        <p>  Total Users To Crawl </p>
                    </div>
                    <div class="row text-center">
                        <p style="font-size: 4.3rem; font-weight: 600;" id="secondCrawlTotalUsersToCrawl" > 0 </p>
                    </div>
                </div>
            </div>

            <div class="row mt-2">
                <div class="col">
                    <div class="row text-center">
                        <p> Status </p>
                    </div>
                    <div class="row text-center">
                        <p style="font-size: 1.9rem; font-weight: 600;" id="secondCrawlCrawlStatus"> Initialising </p>
                    </div>
                </div>
                <div class="col">
                    <div class="row text-center">
                        <p>  Started Crawl </p>
                    </div>
                    <div class="row text-center">
                        <p style="font-size: 1.9rem; font-weight: 600;" id="secondCrawlCrawlTime"> 0s </p>
                    </div>
                </div>
            </div>

            <!-- crawling progress bar -->
            <div class="row crawlProgressBar mt-4">
                <span class="progressBarSpan" id="secondCrawlProgressBarID"></span>
            </div>

            <!-- crawling progress bar -->
            <div class="row text-center mt-1">
                <div class="col" style="font-size: 0.9rem;" id="secondCrawlPercentageDone"> 0% </div>
            </div>
                


            <hr class="mt-3 mb-4">

            <!-- User info card -->
            <div class="row mt-3 mb-3">
                <div class="col box">
                    <div class="row ml-0">
                        <div class="col-8">
                            <div class="row">
                                <div class="col mb-2">
                                    <p id="secondCrawlUserUsername" class="skeleton" style="font-size: 2rem; font-weight: 700; width: 80%; height: 3rem" >
                                        
                                    </p>
                                </div>
                            </div>
                            <div class="row">
                                <div class="col-4">
                                    <p>
                                        Country:
                                    </p>
                                </div>
                                <div class="col">
                                    <p id="secondCrawlUserCountry" style="font-weight: 500" class="skeleton skeleton-text truncate"></p>
                                        
                                    </p>
                                </div>
                            </div>
                            <div class="row">
                                <div class="col-4">
                                    <p>
                                        Friends: 
                                    </p>
                                </div>
                                <div class="col">
                                    <p id="secondCrawlUserFriendCount" style="font-weight: 500; width: 90%" class="skeleton skeleton-text">
                                        
                                    </p>
                                </div>
                            </div>
                            <div class="row">
                                <div class="col-4">
                                    <p>
                                        Created:
                                    </p>
                                </div>
                                <div class="col">
                                    <p id="secondCrawlUserCreationDate" style="font-weight: 500" class="skeleton skeleton-text">
                                        
                                    </p>
                                </div>
                            </div>
                            <div class="row">
                                <div class="col-4">
                                    <p>
                                        Profile: 
                                    </p>
                                </div>
                                <div class="col">
                                    <p id="secondCrawlUserProfile" style="font-weight: 500; width: 83%" class="skeleton skeleton-text">
                                        
                                    </p>
                                </div>
                            </div>
                        </div>
                        
                        <div class="col" id="firstCrawlUserAvatarCol">
                            <div class="secondCrawlUserProfiler" style="position: relative">
                                <img class="mr-1 float-end skeleton" style="border-radius: 10px; height: 10rem; width: 10rem" id="secondCrawlUserAvatar"></img>
                            </div>
                            <div class="secondCrawlUserCountry " style="position: absolute; bottom: 0; left: 12%;">
                                <img id="secondCrawlUserCountryFlag" class="float-end" style="width: 50%"></img>
                            </div>
                        </div>
                    </div>
                </div>
            </div> <!-- end user info card -->
        </div>
    `;
}
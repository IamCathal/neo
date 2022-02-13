import { setCrawlPageUserCardDetails } from '/static/javascript/userCard.js';
import { initAndMonitorCrawlingStatusWebsocket } from '/static/javascript/crawlingStatusUpdatesFeed.js'


const crawlIDs = getCrawlIDs()
console.log(`got crawlIDs: ${crawlIDs}`)

if (crawlIDs.length == 2) {
    getShortestDistanceInfo(crawlIDs).then(shortestDistanceInfo => {
        if (shortestDistanceInfo.crawlids === null) {
            // Does not exist, continue with crawl
            renderCrawlStatusBoxes(2)
            console.log(`opening websockets for ${crawlIDs}`)
            let initAndMonitorCrawlOne = initAndMonitorCrawlingStatusWebsocket(crawlIDs[0], "firstCrawl")
            let initAndMonitorCrawlTwo = initAndMonitorCrawlingStatusWebsocket(crawlIDs[1], "secondCrawl")

            getCrawlingUserWhenAvailable(crawlIDs[0], "firstCrawl").then(res => {}, err => {
                console.error(`error getting first crawling user: ${JSON.stringify(err)}`)
            })
            getCrawlingUserWhenAvailable(crawlIDs[1], "secondCrawl").then(res => {}, err => {
                console.error(`error getting second crawling user: ${JSON.stringify(err)}`)
            })
            Promise.all([initAndMonitorCrawlOne, initAndMonitorCrawlTwo]).then(vals => {
                console.log("both crawls are done!!!!")
                // generate graphs
                let firstUserStartCreateGraph = startCreateGraph(crawlIDs[0])
                let secondUserStartCreateGraph = startCreateGraph(crawlIDs[1])

                Promise.all([firstUserStartCreateGraph, secondUserStartCreateGraph]).then(vals => {
                    // both crawls are complete
                    console.log("both user create graphs have been initialsed")
                    let firstUserGraphExists = waitUntilGraphDataExists(crawlIDs[0])
                    let secondUserGraphExists = waitUntilGraphDataExists(crawlIDs[1])

                    Promise.all([firstUserGraphExists, secondUserGraphExists]).then(vals => {
                        console.log("both users have existing graph data, calculating shortest distance")
                        startCalculateGetShortestDistance(crawlIDs).then(res => {
                            console.log(res);
                            window.location.href = `/shortestdistance?firstcrawlid=${crawlIDs[0]}&secondcrawlid=${crawlIDs[1]}`
                        }, err => {
                            console.error(`err calculating shortest distance: ${JSON.stringify(err)}`)
                        })
                    }, err => {
                        console.error(`error waiting until both graphs existed: ${err}`)
                    })
                }, errs => {
                    console.error(errs)
                })
            }, err => {
                console.error(err)
            })
        } else {
            console.log(`Redirecting to http://localhost:8088/shortestdistance?firstcrawlid=${crawlIDs[0]}&secondcrawlid=${crawlIDs[1]}`)
        }
    }, err => {
        console.error(err)
    })
}

if (crawlIDs.length == 1) {
    doesProcessedGraphDataExist(crawlIDs[0]).then(doesExist => {
        if (doesExist) {
            console.log("did exist")
            // forward to that page
            // window.location.href = `/graph/${crawlID}`;
        } else {
            console.log("no exist")
            renderCrawlStatusBoxes(1)
            // subscribe to crawling status updates
            initAndMonitorCrawlingStatusWebsocket(crawlIDs[0], "firstCrawl").then(res => {
                startCreateGraph(crawlIDs[0]).then(res => {
              
                // Check every 500ms is the graph is done processing yet
                let interval = setInterval(function() {
                    doesProcessedGraphDataExist(crawlIDs[0]).then(doesExist => {
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
            getCrawlingUserWhenAvailable(crawlIDs[0], "firstCrawl").then(res => {}, err => {
                console.error(`error getting crawling user: ${err}`)
            })
        }
    }, err => {
        console.error(`error calling get finished graph data: ${err}`);
    })
}

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

function startCalculateGetShortestDistance(crawlIDs) {
    return new Promise((resolve, reject) => {
        fetch(`http://localhost:2590/api/calculateshortestdistanceinfo`, {
            method: 'POST',
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({"crawlids": crawlIDs})
        }).then((res => res.json()))
        .then(data => {
            resolve(data)
        }).catch(err => {
            reject(err)
        })
    })
}

function startCreateGraph(crawlID) {
    return new Promise((resolve, reject) => {
        fetch(`http://localhost:2570/creategraph/${crawlID}`, {
            method: 'POST',
            headers: {
                "Content-Type": "application/json"
            },
        }).then((res => res.json()))
        .then(data => {
            resolve(data)
        }).catch(err => {
            console.error(err);
        })
    })
}

function getCrawlingUserWhenAvailable(crawlID , idPrefix) {
    return new Promise((resolve, reject) => {
    let interval = setInterval(() => {
        if (usersCrawledIsMoreThanZero(idPrefix)) {
            clearInterval(interval);
            getUser(crawlID).then(user => {
                console.log(`got user`)
                console.log(user)
                setCrawlPageUserCardDetails(user, idPrefix)
                resolve(true)
            }, err => {
                reject(err)
            })
        }
      }, 50);
    });
}

function waitUntilGraphDataExists(crawlID) {
    return new Promise((resolve, reject) => {
        let interval = setInterval(function() {
            doesProcessedGraphDataExist(crawlID).then(doesExist => {
                if (doesExist === true) {
                    clearInterval(interval);
                    resolve(true)
                } else {
                    console.log("graph not done processing")
                }
            }, err => {
                clearInterval(interval);
                reject(err)
            })
        }, 500);
    })
}

function usersCrawledIsMoreThanZero(idPrefix) {
    if (parseInt(document.getElementById(`${idPrefix}UsersCrawled`).textContent) >= 1) {
        return true
    }
    return false
}

function getCrawlIDs() {
    const URLarr = window.location.href.split("/");
    let firstCrawlID = URLarr[URLarr.length-1];

    const queryParams = new URLSearchParams(window.location.search);
    const secondCrawlID = queryParams.get("secondcrawlid")

    if (secondCrawlID != undefined) {
        firstCrawlID = firstCrawlID.split("?")[0]
        return [firstCrawlID, secondCrawlID]
    } else {
        return [firstCrawlID]
    }
}

function getShortestDistanceInfo(crawlIDs) {
    return new Promise((resolve, reject) => {
        console.log("get user")
        fetch(`http://localhost:2590/api/getshortestdistanceinfo`, {
            method: 'POST',
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({"crawlids":crawlIDs})
        }).then((res => res.json()))
        .then(data => {
            if (data.error) {
                reject(data)
            }
            resolve(data.shortestdistanceinfo)
        }).catch(err => {
            console.error(err);
        })
    })
}

function getUser(crawlID) {
    return new Promise((resolve, reject) => {
        console.log("get user")
        fetch(`http://localhost:2590/api/getcrawlinguser/${crawlID}`, {
            headers: {
                "Content-Type": "application/json"
            },
        }).then((res => res.json()))
        .then(data => {
            if (data.error) {
                reject(data)
            }
            resolve(data.user)
        }).catch(err => {
            console.error(err);
        })
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

// TODO COMMON
function countryCodeToName(code) {
    if (countryCodeToNameObj[code] == undefined) {
        return code;
    }
    return countryCodeToNameObj[code]
  }
  // https://gist.github.com/maephisto/9228207
  const countryCodeToNameObj = {
    'AF' : 'Afghanistan',
    'AX' : 'Aland Islands',
    'AL' : 'Albania',
    'DZ' : 'Algeria',
    'AS' : 'American Samoa',
    'AD' : 'Andorra',
    'AO' : 'Angola',
    'AI' : 'Anguilla',
    'AQ' : 'Antarctica',
    'AG' : 'Antigua And Barbuda',
    'AR' : 'Argentina',
    'AM' : 'Armenia',
    'AW' : 'Aruba',
    'AU' : 'Australia',
    'AT' : 'Austria',
    'AZ' : 'Azerbaijan',
    'BS' : 'Bahamas',
    'BH' : 'Bahrain',
    'BD' : 'Bangladesh',
    'BB' : 'Barbados',
    'BY' : 'Belarus',
    'BE' : 'Belgium',
    'BZ' : 'Belize',
    'BJ' : 'Benin',
    'BM' : 'Bermuda',
    'BT' : 'Bhutan',
    'BO' : 'Bolivia',
    'BA' : 'Bosnia And Herzegovina',
    'BW' : 'Botswana',
    'BV' : 'Bouvet Island',
    'BR' : 'Brazil',
    'IO' : 'British Indian Ocean Territory',
    'BN' : 'Brunei Darussalam',
    'BG' : 'Bulgaria',
    'BF' : 'Burkina Faso',
    'BI' : 'Burundi',
    'KH' : 'Cambodia',
    'CM' : 'Cameroon',
    'CA' : 'Canada',
    'CV' : 'Cape Verde',
    'KY' : 'Cayman Islands',
    'CF' : 'Central African Republic',
    'TD' : 'Chad',
    'CL' : 'Chile',
    'CN' : 'China',
    'CX' : 'Christmas Island',
    'CC' : 'Cocos (Keeling) Islands',
    'CO' : 'Colombia',
    'KM' : 'Comoros',
    'CG' : 'Congo',
    'CD' : 'Congo, Democratic Republic',
    'CK' : 'Cook Islands',
    'CR' : 'Costa Rica',
    'CI' : 'Cote D\'Ivoire',
    'HR' : 'Croatia',
    'CU' : 'Cuba',
    'CY' : 'Cyprus',
    'CZ' : 'Czech Republic',
    'DK' : 'Denmark',
    'DJ' : 'Djibouti',
    'DM' : 'Dominica',
    'DO' : 'Dominican Republic',
    'EC' : 'Ecuador',
    'EG' : 'Egypt',
    'SV' : 'El Salvador',
    'GQ' : 'Equatorial Guinea',
    'ER' : 'Eritrea',
    'EE' : 'Estonia',
    'ET' : 'Ethiopia',
    'FK' : 'Falkland Islands (Malvinas)',
    'FO' : 'Faroe Islands',
    'FJ' : 'Fiji',
    'FI' : 'Finland',
    'FR' : 'France',
    'GF' : 'French Guiana',
    'PF' : 'French Polynesia',
    'TF' : 'French Southern Territories',
    'GA' : 'Gabon',
    'GM' : 'Gambia',
    'GE' : 'Georgia',
    'DE' : 'Germany',
    'GH' : 'Ghana',
    'GI' : 'Gibraltar',
    'GR' : 'Greece',
    'GL' : 'Greenland',
    'GD' : 'Grenada',
    'GP' : 'Guadeloupe',
    'GU' : 'Guam',
    'GT' : 'Guatemala',
    'GG' : 'Guernsey',
    'GN' : 'Guinea',
    'GW' : 'Guinea-Bissau',
    'GY' : 'Guyana',
    'HT' : 'Haiti',
    'HM' : 'Heard Island & Mcdonald Islands',
    'VA' : 'Holy See (Vatican City State)',
    'HN' : 'Honduras',
    'HK' : 'Hong Kong',
    'HU' : 'Hungary',
    'IS' : 'Iceland',
    'IN' : 'India',
    'ID' : 'Indonesia',
    'IR' : 'Iran, Islamic Republic Of',
    'IQ' : 'Iraq',
    'IE' : 'Ireland',
    'IM' : 'Isle Of Man',
    'IL' : 'Israel',
    'IT' : 'Italy',
    'JM' : 'Jamaica',
    'JP' : 'Japan',
    'JE' : 'Jersey',
    'JO' : 'Jordan',
    'KZ' : 'Kazakhstan',
    'KE' : 'Kenya',
    'KI' : 'Kiribati',
    'KR' : 'Korea',
    'KW' : 'Kuwait',
    'KG' : 'Kyrgyzstan',
    'LA' : 'Lao People\'s Democratic Republic',
    'LV' : 'Latvia',
    'LB' : 'Lebanon',
    'LS' : 'Lesotho',
    'LR' : 'Liberia',
    'LY' : 'Libyan Arab Jamahiriya',
    'LI' : 'Liechtenstein',
    'LT' : 'Lithuania',
    'LU' : 'Luxembourg',
    'MO' : 'Macao',
    'MK' : 'Macedonia',
    'MG' : 'Madagascar',
    'MW' : 'Malawi',
    'MY' : 'Malaysia',
    'MV' : 'Maldives',
    'ML' : 'Mali',
    'MT' : 'Malta',
    'MH' : 'Marshall Islands',
    'MQ' : 'Martinique',
    'MR' : 'Mauritania',
    'MU' : 'Mauritius',
    'YT' : 'Mayotte',
    'MX' : 'Mexico',
    'FM' : 'Micronesia, Federated States Of',
    'MD' : 'Moldova',
    'MC' : 'Monaco',
    'MN' : 'Mongolia',
    'ME' : 'Montenegro',
    'MS' : 'Montserrat',
    'MA' : 'Morocco',
    'MZ' : 'Mozambique',
    'MM' : 'Myanmar',
    'NA' : 'Namibia',
    'NR' : 'Nauru',
    'NP' : 'Nepal',
    'NL' : 'Netherlands',
    'AN' : 'Netherlands Antilles',
    'NC' : 'New Caledonia',
    'NZ' : 'New Zealand',
    'NI' : 'Nicaragua',
    'NE' : 'Niger',
    'NG' : 'Nigeria',
    'NU' : 'Niue',
    'NF' : 'Norfolk Island',
    'MP' : 'Northern Mariana Islands',
    'NO' : 'Norway',
    'OM' : 'Oman',
    'PK' : 'Pakistan',
    'PW' : 'Palau',
    'PS' : 'Palestinian Territory, Occupied',
    'PA' : 'Panama',
    'PG' : 'Papua New Guinea',
    'PY' : 'Paraguay',
    'PE' : 'Peru',
    'PH' : 'Philippines',
    'PN' : 'Pitcairn',
    'PL' : 'Poland',
    'PT' : 'Portugal',
    'PR' : 'Puerto Rico',
    'QA' : 'Qatar',
    'RE' : 'Reunion',
    'RO' : 'Romania',
    'RU' : 'Russian Federation',
    'RW' : 'Rwanda',
    'BL' : 'Saint Barthelemy',
    'SH' : 'Saint Helena',
    'KN' : 'Saint Kitts And Nevis',
    'LC' : 'Saint Lucia',
    'MF' : 'Saint Martin',
    'PM' : 'Saint Pierre And Miquelon',
    'VC' : 'Saint Vincent And Grenadines',
    'WS' : 'Samoa',
    'SM' : 'San Marino',
    'ST' : 'Sao Tome And Principe',
    'SA' : 'Saudi Arabia',
    'SN' : 'Senegal',
    'RS' : 'Serbia',
    'SC' : 'Seychelles',
    'SL' : 'Sierra Leone',
    'SG' : 'Singapore',
    'SK' : 'Slovakia',
    'SI' : 'Slovenia',
    'SB' : 'Solomon Islands',
    'SO' : 'Somalia',
    'ZA' : 'South Africa',
    'GS' : 'South Georgia And Sandwich Isl.',
    'ES' : 'Spain',
    'LK' : 'Sri Lanka',
    'SD' : 'Sudan',
    'SR' : 'Suriname',
    'SJ' : 'Svalbard And Jan Mayen',
    'SZ' : 'Swaziland',
    'SE' : 'Sweden',
    'CH' : 'Switzerland',
    'SY' : 'Syrian Arab Republic',
    'TW' : 'Taiwan',
    'TJ' : 'Tajikistan',
    'TZ' : 'Tanzania',
    'TH' : 'Thailand',
    'TL' : 'Timor-Leste',
    'TG' : 'Togo',
    'TK' : 'Tokelau',
    'TO' : 'Tonga',
    'TT' : 'Trinidad And Tobago',
    'TN' : 'Tunisia',
    'TR' : 'Turkey',
    'TM' : 'Turkmenistan',
    'TC' : 'Turks And Caicos Islands',
    'TV' : 'Tuvalu',
    'UG' : 'Uganda',
    'UA' : 'Ukraine',
    'AE' : 'United Arab Emirates',
    'GB' : 'United Kingdom',
    'US' : 'United States',
    'UM' : 'United States Outlying Islands',
    'UY' : 'Uruguay',
    'UZ' : 'Uzbekistan',
    'VU' : 'Vanuatu',
    'VE' : 'Venezuela',
    'VN' : 'Viet Nam',
    'VG' : 'Virgin Islands, British',
    'VI' : 'Virgin Islands, U.S.',
    'WF' : 'Wallis And Futuna',
    'EH' : 'Western Sahara',
    'YE' : 'Yemen',
    'ZM' : 'Zambia',
    'ZW' : 'Zimbabwe'
  };

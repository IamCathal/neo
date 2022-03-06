import { countUpElement } from '/static/javascript/countUpScript.js';

const crawlIDs = getCrawlIDsFromShortestDistanceURL()

getShortestDistance(crawlIDs).then(shortestDistanceInfo => {
    console.log(shortestDistanceInfo)

    fillInStatBoxes(shortestDistanceInfo)
    fillInShortestPathMenu(shortestDistanceInfo)
    initEchartsGraph(shortestDistanceInfo)
    fillInIndivCrawlDataBoxes(shortestDistanceInfo)

    initLinkForInteractiveGraphPage()

}, (err) => {
    console.error(err)
})

function fillInStatBoxes(shortestDistanceInfo) {
    countUpElement("networkSpan", shortestDistanceInfo.totalnetworkspan)
    countUpElement("spanLength", shortestDistanceInfo.shortestdistance.length)
    countUpElement("shortestPathCountrySpan", countUniqueCountriesInShortestPath(shortestDistanceInfo.shortestdistance))
    countUpElement("totalDistanceTravelled", getTotalDistanceTravelled(shortestDistanceInfo.shortestdistance), {"suffix": "km"})
}

function countUniqueCountriesInShortestPath(shortestdistancePath) {
    return new Set(getAllCountriesInShortestPath(shortestdistancePath)).size
}

function getTotalDistanceTravelled(shortestDistancePath) {
    const countries = getAllCountriesInShortestPath(shortestDistancePath)
    let totalDistance = 0
    for (let i = 0; i < countries.length - 1; i++) {
        let firstCountryCoords = countryCodeToCoords[countries[i].toUpperCase()]
        let secondCountryCoords = countryCodeToCoords[countries[i+1].toUpperCase()]
        const diff = totalDistance += getDistanceBetweenCoordsInKM(firstCountryCoords, secondCountryCoords)
        console.log(`Distance between ${countries[i].toUpperCase()} and ${countries[i+1].toUpperCase()} is ${diff}km`)
        totalDistance += diff
    }
    return totalDistance
}

function fillInShortestPathMenu(shortestDistanceInfo) {
    document.getElementById("firstUserUsername").textContent = shortestDistanceInfo.firstuser.accdetails.personaname
    document.getElementById("secondUserUsername").textContent = shortestDistanceInfo.seconduser.accdetails.personaname
    shortestDistanceInfo.shortestdistance.forEach(user => {

    let fullResProfiler = user.accdetails.avatar.split(".jpg").join("") + "_full.jpg";
    const creationDate = new Date(user.accdetails.timecreated*1000);
    const dateString = `${creationDate.getDate()} ${creationDate.toLocaleString('default', { month: 'long' })} ${creationDate.getFullYear()}`;
    const timeSinceString = `(${timezSince(creationDate)} ago)`

    document.getElementById("shortestPathDiv").innerHTML += `
    <div class="row pl-3 pr-3 mt-4">
            <div class="col">
                <div class="row">
                    <div class="col-3">
                        <div class="text-center">
                            <a href="${user.accdetails.profileurl}">
                                <img class="mr-1 skeleton" style="border-radius: 10px; height: 10rem; width: 10rem" src="${fullResProfiler}"></img>
                            </a>
                        </div>
                    </div>
                    <div class="col">
                        <div class="row m-0">
                            <p class="truncate mb-0" style="font-size: 3rem; font-weight: 500;">
                                ${user.accdetails.personaname}
                            </p>
                        </div>
                        <div class="row m-0 mt-0">
                            <p class="mb-0">
                                ${user.friendids.length} friends
                            </p>
                        </div>
                        <div class="row m-0">
                            <p class="mb-0">
                                From ${getFlagEmoji(user.accdetails.loccountrycode)}
                            </p>
                        </div>
                        <div class="row m-0">
                            <p class="mb-0">
                                Created: ${dateString} ${timeSinceString}
                            </p>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
    `;
    });
}

function initLinkForInteractiveGraphPage() {
    if (crawlIDs.length == 2) {
        document.getElementById("interactiveGraphLink").href = `/graph/interactive?firstcrawlid=${crawlIDs[0]}&secondcrawlid=${crawlIDs[1]}`
        return
    }
    document.getElementById("interactiveGraphLink").href = `/graph/interactive?firstcrawlid=${crawlIDs[0]}`
}

function getAllCountriesInShortestPath(shortestDistancePath) {
    let countries = []
    shortestDistancePath.forEach(user => {
        console.log(user.accdetails.loccountrycode)
        if (user.accdetails.loccountrycode != "") {
            countries.push(user.accdetails.loccountrycode)
        }
    })
    return countries
}

// COMMON
function timezSince(targetDate) {
    let seconds = Math.floor((new Date()-targetDate)/1000)
    let interval = seconds / 31536000 
    if (interval > 1) {
        return Math.floor(interval) + " years";
    }
    interval = seconds / 2592000; // months
    if (interval > 1) {
        return Math.floor(interval) + " months";
      }
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

function getShortestDistance(bothCrawlIDs) {
    return new Promise((resolve, reject) => {
        fetch(`http://localhost:2590/api/getshortestdistanceinfo`, {
            method: 'POST',
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({"crawlids":bothCrawlIDs})
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


function getCrawlIDsFromShortestDistanceURL() {
    const queryParams = new URLSearchParams(window.location.search);
    const firstCrawlID = queryParams.get("firstcrawlid")
    const secondCrawlID = queryParams.get("secondcrawlid")

    return [firstCrawlID, secondCrawlID]
}

function initEchartsGraph(shortestDistanceData) {
    let myChart = echarts.init(document.getElementById('graphContainer'));
    let nodes = []
    let links = []

    shortestDistanceData.shortestdistance.forEach(user => {
        nodes.push({
            "id": user.accdetails.steamid,
            "name": user.accdetails.personaname,
            "value": user.accdetails.profileurl,
            "symbol": `image://${user.accdetails.avatar}`
        });
    });
    
    for (let i = 0; i < shortestDistanceData.shortestdistance.length - 1; i++) {
        links.push({
            "source": shortestDistanceData.shortestdistance[i].accdetails.steamid,
            "target": shortestDistanceData.shortestdistance[i+1].accdetails.steamid
        })
    }
    console.log(nodes)
    console.log(links)

    var option;
    myChart.showLoading();
    myChart.hideLoading()
    nodes.forEach(function (node) {
        node.symbolSize = 80;
    });
    option = {
        title: {
            text: 'Your friend network',
            subtext: 'Default layout',
            top: 'bottom',
            left: 'right',
            textStyle: {
                color: '#ffffff'
            }
        },

        tooltip: {
            show: true,
            showContent: true,
            triggerOn: 'click',
            enterable: true,
            renderMode: 'html',
            formatter: function(params, ticket, callback) {
                return `<div>
                            <p style="font-weight: bold; color: black" class="tooltipText">${params["name"]}:</p> 
                            <a href="${params["data"].value}" target="_blank">
                                <button class="tooltipButton" style="color: black; background-color: #27b8f0;">Profile</button>
                            </a>
                        </div>`
            }
        },
        legend: [
        {
            show: true,
            left: 'left',
            textStyle: {
                color: '#ffffff'
            }
        }
        ],
        series: [
        {
            name: 'Friend Network',
            type: 'graph',
            layout: 'force',
            data: nodes,
            links: links,
            // categories: graph.categories,
            roam: true,
            draggable: true,
            label: {
                position: 'right'
            },
            force: {
                gravity: 0.2,
                repulsion: 5500,
                friction: 0.2,
            }
        }
        ]
    };
    myChart.setOption(option);
    option && myChart.setOption(option);
}

// COMMON
// https://dev.to/jorik/country-code-to-flag-emoji-a21
function getFlagEmoji(countryCode) {
    const codePoints = countryCode
      .toUpperCase()
      .split('')
      .map(char =>  127397 + char.charCodeAt());
    return String.fromCodePoint(...codePoints);
}

function fillInIndivCrawlDataBoxes(shortestDistanceInfo) {
    document.getElementById("indivCrawlData").innerHTML += `
    <div class="col-5">
        <a href="/graph/${shortestDistanceInfo.crawlids[0]}">
            <div class="row pt-2 pb-2" style="border-radius: 5px">
                <div class="col-2">
                    <img src="${shortestDistanceInfo.firstuser.accdetails.avatar}">
                </div>
                <div class="col">
                    <p style="font-weight: 500; font-size: 1.4rem" class="mb-0">${shortestDistanceInfo.firstuser.accdetails.personaname}</p>
                </div>
            </div>
        </a>
    </div>
    `

    document.getElementById("indivCrawlData").innerHTML += `
    <div class="col-5 ml-4">
        <a href="/graph/${shortestDistanceInfo.crawlids[1]}">
            <div class="row pt-2 pb-2" style="border-radius: 5px">
                <div class="col-2">
                    <img src="${shortestDistanceInfo.seconduser.accdetails.avatar}">
                </div>
                <div class="col">
                    <p style="font-weight: 500; font-size: 1.4rem" class="mb-0">${shortestDistanceInfo.seconduser.accdetails.personaname}</p>
                </div>
            </div>
        </a>
    </div>
    `
}

function getDistanceBetweenCoordsInKM(firstCoord, secondCoord) {
    const radius = 6371 
    const latitudeDiff = toRadians(secondCoord.latitude - firstCoord.latitude)
    const longitudeDiff = toRadians(secondCoord.longitude - firstCoord.longitude)
    let firstCoordLatitude = toRadians(firstCoord.latitude)
    let secondCoordLatitude = toRadians(secondCoord.latitude)

    let a = Math.sin(latitudeDiff/2) * Math.sin(latitudeDiff/2) + Math.sin(longitudeDiff/2) * Math.sin(longitudeDiff/2) * 
        Math.cos(firstCoordLatitude) * Math.cos(secondCoordLatitude)
    let distance = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1-a))
    return radius * distance
}

function toRadians(degrees) {
    return degrees * Math.PI / 180;
}

// COMMON COPIED
const countryCodeToCoords = {
    "AD": {
        "latitude": 42.546245,
        "longitude": 1.601554
    },
    "AE": {
        "latitude": 23.424076,
        "longitude": 53.847818
    },
    "AF": {
        "latitude": 33.93911,
        "longitude": 67.709953
    },
    "AG": {
        "latitude": 17.060816,
        "longitude": -61.796428
    },
    "AI": {
        "latitude": 18.220554,
        "longitude": -63.068615
    },
    "AL": {
        "latitude": 41.153332,
        "longitude": 20.168331
    },
    "AM": {
        "latitude": 40.069099,
        "longitude": 45.038189
    },
    "AN": {
        "latitude": 12.226079,
        "longitude": -69.060087
    },
    "AO": {
        "latitude": -11.202692,
        "longitude": 17.873887
    },
    "AQ": {
        "latitude": -75.250973,
        "longitude": -0.071389
    },
    "AR": {
        "latitude": -38.416097,
        "longitude": -63.616672
    },
    "AS": {
        "latitude": -14.270972,
        "longitude": -170.132217
    },
    "AT": {
        "latitude": 47.516231,
        "longitude": 14.550072
    },
    "AU": {
        "latitude": -25.274398,
        "longitude": 133.775136
    },
    "AW": {
        "latitude": 12.52111,
        "longitude": -69.968338
    },
    "AZ": {
        "latitude": 40.143105,
        "longitude": 47.576927
    },
    "BA": {
        "latitude": 43.915886,
        "longitude": 17.679076
    },
    "BB": {
        "latitude": 13.193887,
        "longitude": -59.543198
    },
    "BD": {
        "latitude": 23.684994,
        "longitude": 90.356331
    },
    "BE": {
        "latitude": 50.503887,
        "longitude": 4.469936
    },
    "BF": {
        "latitude": 12.238333,
        "longitude": -1.561593
    },
    "BG": {
        "latitude": 42.733883,
        "longitude": 25.48583
    },
    "BH": {
        "latitude": 25.930414,
        "longitude": 50.637772
    },
    "BI": {
        "latitude": -3.373056,
        "longitude": 29.918886
    },
    "BJ": {
        "latitude": 9.30769,
        "longitude": 2.315834
    },
    "BM": {
        "latitude": 32.321384,
        "longitude": -64.75737
    },
    "BN": {
        "latitude": 4.535277,
        "longitude": 114.727669
    },
    "BO": {
        "latitude": -16.290154,
        "longitude": -63.588653
    },
    "BR": {
        "latitude": -14.235004,
        "longitude": -51.92528
    },
    "BS": {
        "latitude": 25.03428,
        "longitude": -77.39628
    },
    "BT": {
        "latitude": 27.514162,
        "longitude": 90.433601
    },
    "BV": {
        "latitude": -54.423199,
        "longitude": 3.413194
    },
    "BW": {
        "latitude": -22.328474,
        "longitude": 24.684866
    },
    "BY": {
        "latitude": 53.709807,
        "longitude": 27.953389
    },
    "BZ": {
        "latitude": 17.189877,
        "longitude": -88.49765
    },
    "CA": {
        "latitude": 56.130366,
        "longitude": -106.346771
    },
    "CC": {
        "latitude": -12.164165,
        "longitude": 96.870956
    },
    "CC": {
        "latitude": -12.164165,
        "longitude": 96.870956
    },
    "CD": {
        "latitude": -4.038333,
        "longitude": 21.758664
    },
    "CF": {
        "latitude": 6.611111,
        "longitude": 20.939444
    },
    "CG": {
        "latitude": -0.228021,
        "longitude": 15.827659
    },
    "CH": {
        "latitude": 46.818188,
        "longitude": 8.227512
    },
    "CI": {
        "latitude": 7.539989,
        "longitude": -5.54708
    },
    "CK": {
        "latitude": -21.236736,
        "longitude": -159.777671
    },
    "CL": {
        "latitude": -35.675147,
        "longitude": -71.542969
    },
    "CM": {
        "latitude": 7.369722,
        "longitude": 12.354722
    },
    "CN": {
        "latitude": 35.86166,
        "longitude": 104.195397
    },
    "CO": {
        "latitude": 4.570868,
        "longitude": -74.297333
    },
    "CR": {
        "latitude": 9.748917,
        "longitude": -83.753428
    },
    "CU": {
        "latitude": 21.521757,
        "longitude": -77.781167
    },
    "CV": {
        "latitude": 16.002082,
        "longitude": -24.013197
    },
    "CX": {
        "latitude": -10.447525,
        "longitude": 105.690449
    },
    "CY": {
        "latitude": 35.126413,
        "longitude": 33.429859
    },
    "CZ": {
        "latitude": 49.817492,
        "longitude": 15.472962
    },
    "DE": {
        "latitude": 51.165691,
        "longitude": 10.451526
    },
    "DJ": {
        "latitude": 11.825138,
        "longitude": 42.590275
    },
    "DJ": {
        "latitude": 11.825138,
        "longitude": 42.590275
    },
    "DK": {
        "latitude": 56.26392,
        "longitude": 9.501785
    },
    "DM": {
        "latitude": 15.414999,
        "longitude": -61.370976
    },
    "DO": {
        "latitude": 18.735693,
        "longitude": -70.162651
    },
    "DZ": {
        "latitude": 28.033886,
        "longitude": 1.659626
    },
    "EC": {
        "latitude": -1.831239,
        "longitude": -78.183406
    },
    "EE": {
        "latitude": 58.595272,
        "longitude": 25.013607
    },
    "EG": {
        "latitude": 26.820553,
        "longitude": 30.802498
    },
    "EH": {
        "latitude": 24.215527,
        "longitude": -12.885834
    },
    "ER": {
        "latitude": 15.179384,
        "longitude": 39.782334
    },
    "ES": {
        "latitude": 40.463667,
        "longitude": -3.74922
    },
    "ET": {
        "latitude": 9.145,
        "longitude": 40.489673
    },
    "FI": {
        "latitude": 61.92411,
        "longitude": 25.748151
    },
    "FJ": {
        "latitude": -16.578193,
        "longitude": 179.414413
    },
    "FK": {
        "latitude": -51.796253,
        "longitude": -59.523613
    },
    "FM": {
        "latitude": 7.425554,
        "longitude": 150.550812
    },
    "FO": {
        "latitude": 61.892635,
        "longitude": -6.911806
    },
    "FR": {
        "latitude": 46.227638,
        "longitude": 2.213749
    },
    "GA": {
        "latitude": -0.803689,
        "longitude": 11.609444
    },
    "GB": {
        "latitude": 55.378051,
        "longitude": -3.435973
    },
    "GD": {
        "latitude": 12.262776,
        "longitude": -61.604171
    },
    "GE": {
        "latitude": 42.315407,
        "longitude": 43.356892
    },
    "GF": {
        "latitude": 3.933889,
        "longitude": -53.125782
    },
    "GG": {
        "latitude": 49.465691,
        "longitude": -2.585278
    },
    "GH": {
        "latitude": 7.946527,
        "longitude": -1.023194
    },
    "GI": {
        "latitude": 36.137741,
        "longitude": -5.345374
    },
    "GL": {
        "latitude": 71.706936,
        "longitude": -42.604303
    },
    "GM": {
        "latitude": 13.443182,
        "longitude": -15.310139
    },
    "GN": {
        "latitude": 9.945587,
        "longitude": -9.696645
    },
    "GP": {
        "latitude": 16.995971,
        "longitude": -62.067641
    },
    "GQ": {
        "latitude": 1.650801,
        "longitude": 10.267895
    },
    "GR": {
        "latitude": 39.074208,
        "longitude": 21.824312
    },
    "GS": {
        "latitude": -54.429579,
        "longitude": -36.587909
    },
    "GT": {
        "latitude": 15.783471,
        "longitude": -90.230759
    },
    "GU": {
        "latitude": 13.444304,
        "longitude": 144.793731
    },
    "GW": {
        "latitude": 11.803749,
        "longitude": -15.180413
    },
    "GY": {
        "latitude": 4.860416,
        "longitude": -58.93018
    },
    "GZ": {
        "latitude": 31.354676,
        "longitude": 34.308825
    },
    "HK": {
        "latitude": 22.396428,
        "longitude": 114.109497
    },
    "HM": {
        "latitude": -53.08181,
        "longitude": 73.504158
    },
    "HN": {
        "latitude": 15.199999,
        "longitude": -86.241905
    },
    "HR": {
        "latitude": 45.1,
        "longitude": 15.2
    },
    "HT": {
        "latitude": 18.971187,
        "longitude": -72.285215
    },
    "HU": {
        "latitude": 47.162494,
        "longitude": 19.503304
    },
    "ID": {
        "latitude": -0.789275,
        "longitude": 113.921327
    },
    "IE": {
        "latitude": 53.41291,
        "longitude": -8.24389
    },
    "IL": {
        "latitude": 31.046051,
        "longitude": 34.851612
    },
    "IM": {
        "latitude": 54.236107,
        "longitude": -4.548056
    },
    "IN": {
        "latitude": 20.593684,
        "longitude": 78.96288
    },
    "IO": {
        "latitude": -6.343194,
        "longitude": 71.876519
    },
    "IQ": {
        "latitude": 33.223191,
        "longitude": 43.679291
    },
    "IR": {
        "latitude": 32.427908,
        "longitude": 53.688046
    },
    "IS": {
        "latitude": 64.963051,
        "longitude": -19.020835
    },
    "IT": {
        "latitude": 41.87194,
        "longitude": 12.56738
    },
    "JE": {
        "latitude": 49.214439,
        "longitude": -2.13125
    },
    "JM": {
        "latitude": 18.109581,
        "longitude": -77.297508
    },
    "JO": {
        "latitude": 30.585164,
        "longitude": 36.238414
    },
    "JP": {
        "latitude": 36.204824,
        "longitude": 138.252924
    },
    "KE": {
        "latitude": -0.023559,
        "longitude": 37.906193
    },
    "KG": {
        "latitude": 41.20438,
        "longitude": 74.766098
    },
    "KH": {
        "latitude": 12.565679,
        "longitude": 104.990963
    },
    "KI": {
        "latitude": -3.370417,
        "longitude": -168.734039
    },
    "KM": {
        "latitude": -11.875001,
        "longitude": 43.872219
    },
    "KN": {
        "latitude": 17.357822,
        "longitude": -62.782998
    },
    "KP": {
        "latitude": 40.339852,
        "longitude": 127.510093
    },
    "KR": {
        "latitude": 35.907757,
        "longitude": 127.766922
    },
    "KW": {
        "latitude": 29.31166,
        "longitude": 47.481766
    },
    "KY": {
        "latitude": 19.513469,
        "longitude": -80.566956
    },
    "KZ": {
        "latitude": 48.019573,
        "longitude": 66.923684
    },
    "LA": {
        "latitude": 19.85627,
        "longitude": 102.495496
    },
    "LB": {
        "latitude": 33.854721,
        "longitude": 35.862285
    },
    "LC": {
        "latitude": 13.909444,
        "longitude": -60.978893
    },
    "LI": {
        "latitude": 47.166,
        "longitude": 9.555373
    },
    "LK": {
        "latitude": 7.873054,
        "longitude": 80.771797
    },
    "LR": {
        "latitude": 6.428055,
        "longitude": -9.429499
    },
    "LS": {
        "latitude": -29.609988,
        "longitude": 28.233608
    },
    "LT": {
        "latitude": 55.169438,
        "longitude": 23.881275
    },
    "LU": {
        "latitude": 49.815273,
        "longitude": 6.129583
    },
    "LV": {
        "latitude": 56.879635,
        "longitude": 24.603189
    },
    "LY": {
        "latitude": 26.3351,
        "longitude": 17.228331
    },
    "MA": {
        "latitude": 31.791702,
        "longitude": -7.09262
    },
    "MC": {
        "latitude": 43.750298,
        "longitude": 7.412841
    },
    "MD": {
        "latitude": 47.411631,
        "longitude": 28.369885
    },
    "ME": {
        "latitude": 42.708678,
        "longitude": 19.37439
    },
    "MG": {
        "latitude": -18.766947,
        "longitude": 46.869107
    },
    "MH": {
        "latitude": 7.131474,
        "longitude": 171.184478
    },
    "MK": {
        "latitude": 41.608635,
        "longitude": 21.745275
    },
    "ML": {
        "latitude": 17.570692,
        "longitude": -3.996166
    },
    "MM": {
        "latitude": 21.913965,
        "longitude": 95.956223
    },
    "MN": {
        "latitude": 46.862496,
        "longitude": 103.846656
    },
    "MO": {
        "latitude": 22.198745,
        "longitude": 113.543873
    },
    "MP": {
        "latitude": 17.33083,
        "longitude": 145.38469
    },
    "MQ": {
        "latitude": 14.641528,
        "longitude": -61.024174
    },
    "MR": {
        "latitude": 21.00789,
        "longitude": -10.940835
    },
    "MS": {
        "latitude": 16.742498,
        "longitude": -62.187366
    },
    "MT": {
        "latitude": 35.937496,
        "longitude": 14.375416
    },
    "MU": {
        "latitude": -20.348404,
        "longitude": 57.552152
    },
    "MV": {
        "latitude": 3.202778,
        "longitude": 73.22068
    },
    "MW": {
        "latitude": -13.254308,
        "longitude": 34.301525
    },
    "MX": {
        "latitude": 23.634501,
        "longitude": -102.552784
    },
    "MY": {
        "latitude": 4.210484,
        "longitude": 101.975766
    },
    "MZ": {
        "latitude": -18.665695,
        "longitude": 35.529562
    },
    "NA": {
        "latitude": -22.95764,
        "longitude": 18.49041
    },
    "NC": {
        "latitude": -20.904305,
        "longitude": 165.618042
    },
    "NE": {
        "latitude": 17.607789,
        "longitude": 8.081666
    },
    "NF": {
        "latitude": -29.040835,
        "longitude": 167.954712
    },
    "NG": {
        "latitude": 9.081999,
        "longitude": 8.675277
    },
    "NI": {
        "latitude": 12.865416,
        "longitude": -85.207229
    },
    "NL": {
        "latitude": 52.132633,
        "longitude": 5.291266
    },
    "NO": {
        "latitude": 60.472024,
        "longitude": 8.468946
    },
    "NP": {
        "latitude": 28.394857,
        "longitude": 84.124008
    },
    "NR": {
        "latitude": -0.522778,
        "longitude": 166.931503
    },
    "NU": {
        "latitude": -19.054445,
        "longitude": -169.867233
    },
    "NZ": {
        "latitude": -40.900557,
        "longitude": 174.885971
    },
    "OM": {
        "latitude": 21.512583,
        "longitude": 55.923255
    },
    "PA": {
        "latitude": 8.537981,
        "longitude": -80.782127
    },
    "PE": {
        "latitude": -9.189967,
        "longitude": -75.015152
    },
    "PF": {
        "latitude": -17.679742,
        "longitude": -149.406843
    },
    "PG": {
        "latitude": -6.314993,
        "longitude": 143.95555
    },
    "PH": {
        "latitude": 12.879721,
        "longitude": 121.774017
    },
    "PK": {
        "latitude": 30.375321,
        "longitude": 69.345116
    },
    "PL": {
        "latitude": 51.919438,
        "longitude": 19.145136
    },
    "PM": {
        "latitude": 46.941936,
        "longitude": -56.27111
    },
    "PN": {
        "latitude": -24.703615,
        "longitude": -127.439308
    },
    "PR": {
        "latitude": 18.220833,
        "longitude": -66.590149
    },
    "PS": {
        "latitude": 31.952162,
        "longitude": 35.233154
    },
    "PT": {
        "latitude": 39.399872,
        "longitude": -8.224454
    },
    "PW": {
        "latitude": 7.51498,
        "longitude": 134.58252
    },
    "PY": {
        "latitude": -23.442503,
        "longitude": -58.443832
    },
    "QA": {
        "latitude": 25.354826,
        "longitude": 51.183884
    },
    "RE": {
        "latitude": -21.115141,
        "longitude": 55.536384
    },
    "RO": {
        "latitude": 45.943161,
        "longitude": 24.96676
    },
    "RS": {
        "latitude": 44.016521,
        "longitude": 21.005859
    },
    "RU": {
        "latitude": 61.52401,
        "longitude": 105.31875
    },
    "RW": {
        "latitude": -1.940278,
        "longitude": 29.873888
    },
    "SA": {
        "latitude": 23.885942,
        "longitude": 45.079162
    },
    "SB": {
        "latitude": -9.64571,
        "longitude": 160.156194
    },
    "SC": {
        "latitude": -4.679574,
        "longitude": 55.491977
    },
    "SD": {
        "latitude": 12.862807,
        "longitude": 30.217636
    },
    "SE": {
        "latitude": 60.128161,
        "longitude": 18.643501
    },
    "SG": {
        "latitude": 1.352083,
        "longitude": 103.819836
    },
    "SH": {
        "latitude": -24.143474,
        "longitude": -10.030696
    },
    "SI": {
        "latitude": 46.151241,
        "longitude": 14.995463
    },
    "SJ": {
        "latitude": 77.553604,
        "longitude": 23.670272
    },
    "SK": {
        "latitude": 48.669026,
        "longitude": 19.699024
    },
    "SL": {
        "latitude": 8.460555,
        "longitude": -11.779889
    },
    "SM": {
        "latitude": 43.94236,
        "longitude": 12.457777
    },
    "SN": {
        "latitude": 14.497401,
        "longitude": -14.452362
    },
    "SO": {
        "latitude": 5.152149,
        "longitude": 46.199616
    },
    "SR": {
        "latitude": 3.919305,
        "longitude": -56.027783
    },
    "ST": {
        "latitude": 0.18636,
        "longitude": 6.613081
    },
    "SV": {
        "latitude": 13.794185,
        "longitude": -88.89653
    },
    "SY": {
        "latitude": 34.802075,
        "longitude": 38.996815
    },
    "SZ": {
        "latitude": -26.522503,
        "longitude": 31.465866
    },
    "TC": {
        "latitude": 21.694025,
        "longitude": -71.797928
    },
    "TD": {
        "latitude": 15.454166,
        "longitude": 18.732207
    },
    "TF": {
        "latitude": -49.280366,
        "longitude": 69.348557
    },
    "TG": {
        "latitude": 8.619543,
        "longitude": 0.824782
    },
    "TH": {
        "latitude": 15.870032,
        "longitude": 100.992541
    },
    "TJ": {
        "latitude": 38.861034,
        "longitude": 71.276093
    },
    "TJ": {
        "latitude": 38.861034,
        "longitude": 71.276093
    },
    "TK": {
        "latitude": -8.967363,
        "longitude": -171.855881
    },
    "TL": {
        "latitude": -8.874217,
        "longitude": 125.727539
    },
    "TM": {
        "latitude": 38.969719,
        "longitude": 59.556278
    },
    "TN": {
        "latitude": 33.886917,
        "longitude": 9.537499
    },
    "TO": {
        "latitude": -21.178986,
        "longitude": -175.198242
    },
    "TR": {
        "latitude": 38.963745,
        "longitude": 35.243322
    },
    "TT": {
        "latitude": 10.691803,
        "longitude": -61.222503
    },
    "TV": {
        "latitude": -7.109535,
        "longitude": 177.64933
    },
    "TW": {
        "latitude": 23.69781,
        "longitude": 120.960515
    },
    "TZ": {
        "latitude": -6.369028,
        "longitude": 34.888822
    },
    "UA": {
        "latitude": 48.379433,
        "longitude": 31.16558
    },
    "UG": {
        "latitude": 1.373333,
        "longitude": 32.290275
    },
    "US": {
        "latitude": 37.09024,
        "longitude": -95.712891
    },
    "UY": {
        "latitude": -32.522779,
        "longitude": -55.765835
    },
    "UZ": {
        "latitude": 41.377491,
        "longitude": 64.585262
    },
    "VA": {
        "latitude": 41.902916,
        "longitude": 12.453389
    },
    "VC": {
        "latitude": 12.984305,
        "longitude": -61.287228
    },
    "VE": {
        "latitude": 6.42375,
        "longitude": -66.58973
    },
    "VG": {
        "latitude": 18.420695,
        "longitude": -64.639968
    },
    "VI": {
        "latitude": 18.335765,
        "longitude": -64.896335
    },
    "VN": {
        "latitude": 14.058324,
        "longitude": 108.277199
    },
    "VU": {
        "latitude": -15.376706,
        "longitude": 166.959158
    },
    "WF": {
        "latitude": -13.768752,
        "longitude": -177.156097
    },
    "WS": {
        "latitude": -13.759029,
        "longitude": -172.104629
    },
    "XK": {
        "latitude": 42.602636,
        "longitude": 20.902977
    },
    "YE": {
        "latitude": 15.552727,
        "longitude": 48.51638
    },
    "YT": {
        "latitude": -12.8275,
        "longitude": 45.166244
    },
    "ZA": {
        "latitude": -30.559482,
        "longitude": 22.937506
    },
    "ZM": {
        "latitude": -13.133897,
        "longitude": 27.849332
    },
    "ZW": {
        "latitude": -19.015438,
        "longitude": 29.154857
    }
}
import { countUpElement } from '/static/javascript/countUpScript.js';


const crawlIDs = getCrawlIDsFromShortestDistanceURL()

getShortestDistance(crawlIDs).then(shortestDistanceInfo => {
    console.log(shortestDistanceInfo)

    fillInStatBoxes(shortestDistanceInfo)
    fillInShortestPathMenu(shortestDistanceInfo)
    initEchartsGraph(shortestDistanceInfo)
    fillInIndivCrawlDataBoxes(shortestDistanceInfo)

}, (err) => {
    console.error(err)
})

function fillInStatBoxes(shortestDistanceInfo) {
    countUpElement("networkSpan", shortestDistanceInfo.totalnetworkspan)
    countUpElement("spanLength", shortestDistanceInfo.shortestdistance.length)
}

function fillInShortestPathMenu(shortestDistanceInfo) {
    document.getElementById("firstUserUsername").textContent = shortestDistanceInfo.firstuser.accdetails.personaname
    document.getElementById("secondUserUsername").textContent = shortestDistanceInfo.seconduser.accdetails.personaname
    shortestDistanceInfo.shortestdistance.forEach(user => {
        let fullResProfiler = user.accdetails.avatar.split(".jpg").join("") + "_full.jpg";
        document.getElementById("shortestPathDiv").innerHTML += `
        <div class="row pl-3 pr-3 mt-4">
                <div class="col">
                    <div class="row">
                        <div class="col-3">
                            <div class="text-center">
                                <img class="mr-1 skeleton" style="border-radius: 10px; height: 10rem; width: 10rem" src="${fullResProfiler}"></img>
                            </div>
                        </div>
                        <div class="col">
                            <div class="row m-0">
                                <p class="truncate mb-0" style="font-size: 3rem; font-weight: 500;">
                                    ${user.accdetails.personaname}
                                </p>
                            </div>
                            <div class="row m-0 mt-2">
                                <p class="skeleton skeleton-text mb-1"></p>
                            </div>
                            <div class="row m-0">
                                <p class="skeleton skeleton-text mb-1"></p>
                            </div>
                            <div class="row m-0">
                                <p class="skeleton skeleton-text mb-1"></p>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
        `;
    });
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
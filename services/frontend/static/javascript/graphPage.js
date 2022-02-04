import { countUpElement } from '/static/javascript/countUpScript.js';
import { setUserCardDetails, timezSince } from '/static/javascript/userCard.js';
import { getHeatmapData, getMaxMonthFrequency } from './heatMapCalendarHelper.js';

const URLarr = window.location.href.split("/");
const crawlID = URLarr[URLarr.length-1];
let crawlData = {}

doesProcessedGraphDataExistz(crawlID).then(doesExist => {
    if (doesExist === false) {
        window.location.href = "/"
    }
    getProcessedGraphData(crawlID).then(crawlDataObj => {
        crawlData = crawlDataObj
        setUserCardDetails(crawlData.usergraphdata.userdetails);
        let countryFrequencies = {}
        var countryFrequenciesArr = []

        crawlDataObj.usergraphdata.frienddetails.forEach(friend => {
            const countryCode = friend.User.accdetails.loccountrycode;
            if (countryCode != "") {
                countryFrequencies[countryCode.toLowerCase()] = countryFrequencies[countryCode.toLowerCase()] ? countryFrequencies[countryCode.toLowerCase()] + 1 : 1;
            }
        });
        countryFrequenciesArr = Object.entries(countryFrequencies)

        // Geographic Stats
        initWorldMap(countryFrequenciesArr)
        fillInFlagsDiv(crawlDataObj.usergraphdata.frienddetails)
        fillInTopStatBoxes(crawlData, countryFrequencies)
        fillInTop10Countries(countryFrequencies)
        fillInFromYourCountryStatBox(crawlDataObj, countryFrequencies)
        fillInContinentCoverage(countryFrequencies)

        // Games stats
        initAndRenderGamesBarChart(getDataForGamesBarChart(crawlDataObj.usergraphdata))
        fillInGamesStatBoxes(crawlDataObj.usergraphdata)

        // Friend network stats
        userCreatedGraph(crawlDataObj.usergraphdata)
        userCreatedMonthChart(crawlDataObj.usergraphdata)
        fillInOldestAndNewestUserCards(crawlDataObj.usergraphdata)
        initAndRenderAccountAgeVsFriendCountChart(crawlDataObj.usergraphdata)

        // Three JS bottom test graph
        initThreeJSGraph(crawlDataObj.usergraphdata)

        var myChart = echarts.init(document.getElementById('graphContainer'));
        const graph = getDataInGraphFormat(crawlDataObj.usergraphdata)
        var option;
        myChart.showLoading();
        myChart.hideLoading()
        graph.nodes.forEach(function (node) {
            node.symbolSize = 10;
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
                                <p style="font-weight: bold" class="tooltipText">${params["name"]}:</p> 
                                <a href="${params["data"].value}" target="_blank">
                                    <button class="tooltipButton">Profile</button>
                                </a>
                            </div>`
                }
            },
            legend: [
            {
                // selectedMode: 'single',
                data: graph.categories.map(function (a) {
                    return a.name;
                })
            }
            ],
            series: [
            {
                name: 'Friend Network',
                type: 'graph',
                layout: 'force',
                data: graph.nodes,
                links: graph.links,
                categories: graph.categories,
                roam: true,
                label: {
                    position: 'right'
                },
                force: {
                    gravity: 0.5,
                    repulsion: 370,
                    friction: 0.2,
                }
            }
            ]
        };
        myChart.setOption(option);
        option && myChart.setOption(option);


                }, err => {
                    console.error(`error retrieving processed graph data: ${err}`)
                })
            }, err => {
                console.error(`error calling does processed graphdata exist: ${err}`)
            })


function getProcessedGraphData(crawlID) {
    return new Promise((resolve, reject) => {
        fetch(`http://localhost:2590/api/getprocessedgraphdata/${crawlID}`, {
            method: "POST",
        }).then(res => res.json())
        .then(data => {
            resolve(data)
        }).catch(err => {
            reject(err)
        })
    });
}

function getDataInGraphFormat(gData) {
    let nodes = []
    let links = []
    nodes.push({
        "id": gData.userdetails.User.accdetails.steamid,
        "name": gData.userdetails.User.accdetails.personaname,
        "value": gData.userdetails.User.accdetails.profileurl,
        "category": 0,
        "symbol": `image://${gData.userdetails.User.accdetails.avatar}`
    })
    gData.frienddetails.forEach((friend) => {
        nodes.push({
            "id": friend.User.accdetails.steamid,
            "name": friend.User.accdetails.personaname,
            "value": friend.User.accdetails.profileurl,
            "category": 0,
            "symbol": `image://${friend.User.accdetails.avatar}`
        });
    });

    if (gData.userdetails.User.friendids.length <= 500) {
        gData.userdetails.User.friendids.forEach(ID => {
            links.push({"source":gData.userdetails.User.accdetails.steamid, "target": `${ID}`})
        })
        gData.frienddetails.forEach(friend => {
            friend.User.friendids.forEach(friendID => {
                links.push({"source":friend.User.accdetails.steamid, "target": `${friendID}`})
            })
        })
    }

    const echartsData = {
        "nodes": nodes,
        "links": links,
        "categories": [{"name": "A"}]
    }
    return echartsData
}

function getDataForGamesBarChart(gData) {
    let yourPlaytimeForTopGames = []
    let networkAveragePlaytimeForTopGames = []
    let topFriendPlaytimeForTopGames = []
    const topOverallGameIDs = gData.topgamedetails.map(game => game.appid)

    // Get your playtime for each of the top games
    for (let i = 0; i < topOverallGameIDs.length; i++) {
        const topCurrGameID = topOverallGameIDs[i]
        let userPlayTimeOfCurrTopGame = 0
        if (i >= gData.userdetails.User.gamesowned.length) {
            yourPlaytimeForTopGames.push(userPlayTimeOfCurrTopGame)
            continue
        }

        gData.userdetails.User.gamesowned.forEach(game => {
            if (topCurrGameID === game.appid) {
                userPlayTimeOfCurrTopGame = game.playtime_forever;
            }
        })        
        yourPlaytimeForTopGames.push(userPlayTimeOfCurrTopGame)
    }

    // Get your networks average and max for each of the top games
    for (let i = 0; i < topOverallGameIDs.length; i++) {
        let topPlaytimeForCurrTopGame = 0
        let totalPlaytimeForThisGame = 0
        let usersWhoHavePlayedThisGame = 0

        gData.frienddetails.forEach(friendObj => {
            const friend = friendObj.User;
            for (let k = 0; k < friend.gamesowned.length; k++) {
                if (friend.gamesowned[k].appid === topOverallGameIDs[i]) {
                    if (friend.gamesowned[k].playtime_forever > topPlaytimeForCurrTopGame) {
                        topPlaytimeForCurrTopGame = friend.gamesowned[k].playtime_forever
                    }
                    totalPlaytimeForThisGame += friend.gamesowned[k].playtime_forever
                    usersWhoHavePlayedThisGame += 1
                }
            }
        })
        const averageNetworkPlaytime = Math.floor(totalPlaytimeForThisGame/usersWhoHavePlayedThisGame)
        networkAveragePlaytimeForTopGames.push(averageNetworkPlaytime)
        topFriendPlaytimeForTopGames.push(topPlaytimeForCurrTopGame)
    }

    // Turn into hours instead of minutes
    yourPlaytimeForTopGames = yourPlaytimeForTopGames.map(minutes => Math.floor(minutes/60))
    networkAveragePlaytimeForTopGames = networkAveragePlaytimeForTopGames.map(minutes => Math.floor(minutes/60))
    topFriendPlaytimeForTopGames = topFriendPlaytimeForTopGames.map(minutes => Math.floor(minutes/60))
    
    const barChartData = {
        xAxisData: gData.topgamedetails.map(game => game.name),
        legend: {
            data: ['You', 'Network Average', 'Top Friend'],
            textStyle: {
                color: '#ffffff'
            }
        },
        yourPlayTimeForTopGamesSeriesObj: {
            name: "You",
            type: 'bar',
            barGap: 0,
            label: "eee",
            emphasis: {
                focus: 'series'
            },
            data: yourPlaytimeForTopGames
        },
        averageNetworkPlayTimeForGameSeriesObj: {
            name: "Network Average",
            type: 'bar',
            barGap: 0,
            label: "eee",
            emphasis: {
                focus: 'series'
            },
            data: networkAveragePlaytimeForTopGames
        },
        topFriendPlayTimeForGameSeriesObj: {
            name: "Top Friend",
            type: 'bar',
            barGap: 0,
            label: "eee",
            emphasis: {
                focus: 'series'
            },
            data: topFriendPlaytimeForTopGames
        }
    }
    return barChartData;
}

function fillInGamesStatBoxes(graphData) {
    console.log(graphData)
    const minWage = 10.20
    let totalHoursPlayedForUser = 0
    graphData.userdetails.User.gamesowned.forEach(game => {
        totalHoursPlayedForUser += game.playtime_forever
    })
    totalHoursPlayedForUser = Math.floor(totalHoursPlayedForUser/60)

    const entireDaysOfPlaytime = Math.floor(totalHoursPlayedForUser / 24)
    const minWageEarnedForGaming = Math.floor(totalHoursPlayedForUser * minWage)
    let percentageOfFriendsWithLessHoursPlayed = 0
    let friendsWithLessHoursPlayed = 0
    if (totalHoursPlayedForUser != 0) {
        graphData.frienddetails.forEach(user => {
            let hoursPlayedForCurrUser = 0
            user.User.gamesowned.forEach(game => {
                hoursPlayedForCurrUser += Math.floor(game.playtime_forever / 60)
            });
            if (hoursPlayedForCurrUser < totalHoursPlayedForUser) {
                friendsWithLessHoursPlayed++
            }
        })
        
        percentageOfFriendsWithLessHoursPlayed = Math.floor(friendsWithLessHoursPlayed / graphData.frienddetails.length)
    }

    countUpElement("statBoxHoursAcrossLibrary", totalHoursPlayedForUser)
    countUpElement("statBoxEntireDaysOfPlaytime", entireDaysOfPlaytime)
    countUpElement("statBoxFriendsWithLessHoursPlayed", percentageOfFriendsWithLessHoursPlayed, {suffix: "%"})
    countUpElement("statBoxMinWageEarned", minWageEarnedForGaming, {prefix: "€"})

    removeSkeletonClasses(["statBoxHoursAcrossLibrary", "statBoxEntireDaysOfPlaytime",
            "statBoxMinWageEarned", "statBoxFriendsWithLessHoursPlayed"])
}

function initAndRenderAccountAgeVsFriendCountChart(graphData) {
    console.log(graphData)
    let scatterPlotData = []
    let maxAccountAge = 0

    graphData.frienddetails.forEach(user => {
        const friends = user.User.friendids.length;
        const accAge = user.User.accdetails.timecreated;
        let monthsSinceCreation = monthsSince(accAge)
        if (monthsSinceCreation > maxAccountAge) {
            maxAccountAge = monthsSinceCreation
        }
        scatterPlotData.push([
            friends, monthsSinceCreation
        ])
    })
    console.log(scatterPlotData)
    let chartDom = document.getElementById('accountAgeVsFriendCountScatterPlot');
    let myChart = echarts.init(chartDom);
    let option;

    option = {
    xAxis: {
        axisLabel: {
            formatter: '{value} friends'
        },
        axisLine: {
            lineStyle: {
                color: '#ffffff'
            }
        }
    },
    yAxis: {
        axisLabel: {
            formatter: '{value} months'
        },
        axisLine: {
            lineStyle: {
                color: '#ffffff'
            }
        }
    },
    legend: {
        axisLine: {
            lineStyle: {
                color: '#ffffff'
            }
        }
    },
    visualMap: {
        min: 0,
        max: maxAccountAge,
        inRange: {
            color: ['#f2c31a', '#24b7f2']
        },
        calculable: true,
        textStyle: {
            color: '#ffffff'
        },
        orient: 'vertical',
        right: 10,
        top: 'center'
    },
    tooltip: {
        trigger: 'item',
        axisPointer: {
            type: 'cross',
            label: {
                color: 'black'
            }
        }
    },
    series: [
        {
        symbolSize: 20,
        data: scatterPlotData,
        type: 'scatter'
        }
    ]
    };

    option && myChart.setOption(option);
}

function monthsSince(timestamp) {
    const timeObj = new Date(timestamp * 1000)
    let monthDiff;
    const currTime = new Date();
    monthDiff = (currTime.getFullYear() - timeObj.getFullYear()) * 12
    monthDiff += currTime.getMonth()
    monthDiff -= timeObj.getMonth()
    return monthDiff <= 0 ? 0 : monthDiff
}

function initAndRenderGamesBarChart(barChartData) {
    let app = {};

    let chartDom = document.getElementById('gamesBarChartContainer');
    let myChart = echarts.init(chartDom);
    let option;

    const posList = [
        'left',
        'right',
        'top',
        'bottom',
        'inside',
        'insideTop',
        'insideLeft',
        'insideRight',
        'insideBottom',
        'insideTopLeft',
        'insideTopRight',
        'insideBottomLeft',
        'insideBottomRight'
    ];
    app.configParameters = {
        rotate: {
            min: -90,
            max: 90
        },
        align: {
            options: {
                left: 'left',
                center: 'center',
                right: 'right'
            }
        },
        verticalAlign: {
            options: {
                top: 'top',
                middle: 'middle',
                bottom: 'bottom'
            }
        },
        position: {
            options: posList.reduce(function (map, pos) {
                map[pos] = pos;
                return map;
            }, {})
        },
        distance: {
            min: 0,
            max: 100
        }
    };
    app.config = {
        rotate: 90,
        align: 'left',
        verticalAlign: 'middle',
        position: 'insideBottom',
        distance: 15,
        onChange: function () {
            const labelOption = {
                rotate: app.config.rotate,
                align: app.config.align,
                verticalAlign: app.config.verticalAlign,
                position: app.config.position,
                distance: app.config.distance
            };
            myChart.setOption({
            series: [
                { label: labelOption },
                { label: labelOption },
                { label: labelOption },
                { label: labelOption }
            ]
            });
        }
    };
    const labelOption = {
        show: true,
        position: app.config.position,
        distance: app.config.distance,
        align: app.config.align,
        verticalAlign: app.config.verticalAlign,
        rotate: app.config.rotate,
        formatter: '{c}  {name|{a}}',
        fontSize: 16,
        rich: {
            name: {}
        }
    };
    option = {
        tooltip: {
            trigger: 'axis',
            axisPointer: {
            type: 'shadow'
            },
            textStyle: {
                color: 'black'
            }
        },
        legend: barChartData.legend,
        toolbox: {
            show: true,
            orient: 'vertical',
            left: 'right',
            top: 'center',
            feature: {
            mark: { show: true },
            dataView: { show: true, readOnly: false },
            magicType: { show: true, type: ['line', 'bar', 'stack'] },
            restore: { show: true },
            saveAsImage: { show: true }
            }
        },
        xAxis: [{
            type: 'category',
            axisTick: { show: false },
            data: barChartData.xAxisData,
            axisLine: {
                lineStyle: {
                    color: '#ffffff'
                }
            }
        }],
        yAxis: [{
            type: 'value',
            axisLine: {
                lineStyle: {
                    color: '#ffffff'
                }
            }
        }],
        series: [
            barChartData.yourPlayTimeForTopGamesSeriesObj,
            barChartData.averageNetworkPlayTimeForGameSeriesObj,
            barChartData.topFriendPlayTimeForGameSeriesObj
        ]
    };
    option && myChart.setOption(option);
}

function initThreeJSGraph(crawlData) {
    let seenNodes = new Map()
    let nodes = []
    let links = []

    nodes.push({
        "id": crawlData.userdetails.User.accdetails.steamid, 
        "username": crawlData.userdetails.User.accdetails.personaname,
        "avatar":crawlData.userdetails.User.accdetails.avatar
    })
    seenNodes.set(crawlData.userdetails.User.accdetails.steamid, true)

    crawlData.frienddetails.forEach(friend => {
        nodes.push({
            "id":friend.User.accdetails.steamid, 
            "username": friend.User.accdetails.personaname,
            "avatar":friend.User.accdetails.avatar
        })
        seenNodes.set(friend.User.accdetails.steamid, true)
    })

    crawlData.userdetails.User.friendids.forEach(ID => {
        links.push({
            "source": crawlData.userdetails.User.accdetails.steamid,
            "target": ID
        })
    })
    crawlData.frienddetails.forEach(friend => {
        friend.User.friendids.forEach(ID => {
            if (seenNodes.has(ID)) {
                links.push({
                    "source": friend.User.accdetails.steamid,
                    "target": ID
                })
            }
        })
    })

    links.forEach(link => {
        const src = nodes.filter(node => node.id === link.source)[0];
        const dst = nodes.filter(node => node.id === link.target)[0];

        if (src.neighbourNodes === undefined) {
            src.neighbourNodes = []
        }
        if (dst.neighbourNodes === undefined) {
            dst.neighbourNodes = []
        }
        if (src.neighbourLinks === undefined) {
            src.neighbourLinks = []
        }
        if (dst.neighbourLinks === undefined) {
            dst.neighbourLinks = []
        }
        src.neighbourNodes.push(dst)
        dst.neighbourNodes.push(src)
        src.neighbourLinks.push(link)
        dst.neighbourLinks.push(link)
    });

    const threeJSGraphData = {
        nodes: nodes,
        links: links
    }

    const threeJSGraphDiv = document.getElementById('3d-graph');
    let hoveredNode = null;
    let highlightedNodes = new Set()
    let highlightedLinks = new Set()
    const g = ForceGraph3D()(threeJSGraphDiv)
        .graphData(threeJSGraphData)
        .nodeAutoColorBy('user')
        .nodeThreeObject(({ avatar }) => {
            const imgTexture = new THREE.TextureLoader().load(avatar);
            const material = new THREE.SpriteMaterial({ map: imgTexture });
            const sprite = new THREE.Sprite(material);
            sprite.scale.set(16, 16);
            return sprite;
        })
        .nodeLabel(node => `${node.username}: ${node.id}`)
        .onNodeClick(node => {
            const distance = 90;
            const distRatio = 1 + distance/Math.hypot(node.x, node.y, node.z);

            g.cameraPosition(
                { x: node.x * distRatio, y: node.y * distRatio, z: node.z * distRatio },
                node, 
                3000
            );
            setTimeout(() => {
                window.open(`https://steamcommunity.com/profiles/${node.id}`, '_blank')
            }, 3300)
        })
        .linkWidth(link => highlightedLinks.has(link) ? 4 : 1)
        .linkColor(link => highlightedLinks.has(link) ? 'green' : 'white')
        .linkDirectionalParticles(link => highlightedLinks.has(link) ? 8 : 0)
        .linkDirectionalParticleWidth(3)
        .linkDirectionalParticleColor(() => 'green')
        .onNodeHover(node => {
            if ((!node && !highlightedNodes.size) || (node && hoveredNode === node)) {
                return;
            }

            highlightedLinks.clear()
            highlightedNodes.clear()
            if (node != undefined && node != false) {
                highlightedNodes.add(node)
                node.neighbourNodes.forEach(neighourNode => {
                    highlightedNodes.add(neighourNode);
                });
                node.neighbourLinks.forEach(neighbourLink => {
                    highlightedLinks.add(neighbourLink)
                })
            }

            hoveredNode = node || null;

            g.nodeColor(g.nodeColor())
                .linkWidth(g.linkWidth())
                .linkDirectionalParticles(g.linkDirectionalParticles());
        });

    const linkForce = g
    .d3Force("link")
    .distance(link => {
        return 80 + link.source.neighbourNodes.length;
    });
}

// COMMON
function timezSince2(targetDate) {
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

// COMMON
function doesProcessedGraphDataExistz(crawlID) {
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

function initWorldMap(countriesData) {
    Highcharts.mapChart('firstChart', {
        chart: {
            map: 'custom/world'
        },
        title: {
            text: ''
        },
        mapNavigation: {
            enabled: true,
            buttonOptions: {
                verticalAlign: 'bottom'
            }
        },
        colorAxis: {
            min: 0,
            stops: [
                [0, '#b5f2b3'],
                [0.5, "#0c5c0a"],
                [1, "#065e03"]
            ]
        },
        series: [{
            data: countriesData,
            name: 'Random data',
            states: {
                hover: {
                    color: '#2cb851'
                }
            },
        }]
    });
}

function fillInFlagsDiv(friends) {
    let uniqueCountryCode = extractUniqueCountryCodesFromFriends(friends)
    let i = 0;
    uniqueCountryCode.forEach(countryCode => {
        if (i == 48) {
            return
        }
        document.getElementById("allFlagsDiv").innerHTML += `
        <div class="col-1">
            <p style="font-size: 1.7rem">${getFlagEmoji(countryCode)}</p>
        </div>
        `;
        i++;
    });
}

function fillInTopStatBoxes(graphData, countryFreqs) {
    const UNCountries = 195;
    let uniqueCountryCodes = extractUniqueCountryCodesFromFriends(graphData.usergraphdata.frienddetails)

    countUpElement('statBoxFriendCount', graphData.usergraphdata.userdetails.User.friendids.length)
    countUpElement('statBoxUniqueCountries', uniqueCountryCodes.length)
    countUpElement('statBoxGlobalCoverage', Math.floor((uniqueCountryCodes.length/UNCountries)*100), {suffix: "%"})
    countUpElement('statBoxDictatorships', ruledByDictatorCountries(uniqueCountryCodes))
    countUpElement('statBoxContinentCoverage', getContinentCoverage(countryFreqs), {suffix: "%"})

    removeSkeletonClasses(["statBoxFriendCount", "statBoxUniqueCountries", 
            "statBoxGlobalCoverage", "statBoxDictatorships", "statBoxContinentCoverage"])
}

function fillInFromYourCountryStatBox(graphDataObj, countryFreq) {
    let alsoFromCountry = 0;
    const usersCountry = graphDataObj.usergraphdata.userdetails.User.accdetails.loccountrycode;
    if (usersCountry === undefined) {
        document.getElementById("statBoxAlsoFromYourCountry").textContent = alsoFromCountry;
        removeSkeletonClasses(["statBoxAlsoFromYourCountry"])
        return
    }
    const otherFriendsInUsersCountry = countryFreq[usersCountry.toLowerCase()];
    if (!isNaN(otherFriendsInUsersCountry)) {
        document.getElementById("statBoxAlsoFromYourCountry").textContent = otherFriendsInUsersCountry;
        removeSkeletonClasses(["statBoxAlsoFromYourCountry"])
        return
    }
    document.getElementById("statBoxAlsoFromYourCountry").textContent = alsoFromCountry;
    removeSkeletonClasses(["statBoxAlsoFromYourCountry"])
}

function fillInTop10Countries(countriesFreq) {
    const topTenCountryNames = getTopTenCountries(countriesFreq);
    let i = 1;
    topTenCountryNames.forEach(countryName => {
        document.getElementById("topTenCountriesList").innerHTML += `
            <div class="row ml-1 mr-1" >
                <div class="col-1">
                    <p class="gameLeaderBoardText">${i}.</p> 
                </div>
                <div class="col">
                    <p class="gameLeaderBoardText">${countryName}</p>
                </div>  
            </div>
        `;
        i++;
    });
}

function userCreatedGraph(graphData) {
    let friendCreationDatesByMonth = {}
    let creationDates = []
    let creationTimestamps = []
    let creationYearFrequencies = {}
    
    graphData.frienddetails.forEach(friend => {
        creationTimestamps.push(friend.User.accdetails.timecreated)
    })
    creationTimestamps.sort(function(j, k) {
        return j < k;
    })

    creationTimestamps.forEach(date => {
        creationDates.push(new Date(date * 1000))
    })
    
    creationDates.forEach(date => {
        creationYearFrequencies[date.getFullYear()] = creationYearFrequencies[date.getFullYear()] ? creationYearFrequencies[date.getFullYear()] + 1 : 1;
    })

    const oldestYear = Object.entries(creationYearFrequencies)[0][0]
    for (let i = oldestYear; i < new Date().getFullYear(); i++) {
        creationYearFrequencies[i] = creationYearFrequencies[i] ? creationYearFrequencies[i] : 0
    }

    let chartDom = document.getElementById('creationYearBarChartContainer');
    let myChart = echarts.init(chartDom);
    let option;

    option = {
    xAxis: {
        type: 'category',
        data: Object.keys(creationYearFrequencies),
        axisLine: {
            lineStyle: {
                color: '#ffffff'
            }
        }
    },
    yAxis: {
        type: 'value',
        axisLine: {
            lineStyle: {
                color: '#ffffff'
            }
        }
    },
    series: [
        {
        data: Object.values(creationYearFrequencies),
        type: 'bar'
        }
    ]
    };
    option && myChart.setOption(option);
}

function userCreatedMonthChart(graphData) {
    let creationDates = []
    let creationTimestamps = []
    
    graphData.frienddetails.forEach(friend => {
        creationTimestamps.push(friend.User.accdetails.timecreated)
    })
    creationTimestamps.sort(function(j, k) {
        return j < k;
    })

    creationTimestamps.forEach(date => {
        creationDates.push(new Date(date * 1000))
    })

    let chartDom = document.getElementById('creationMonthHeatMapContainer');
    let myChart = echarts.init(chartDom);
    let option;

    let heatmapData = getHeatmapData(creationDates)
    fillInMonthStatBoxes(creationDates)

    option = {
        visualMap: {
            show: false,
            min: 0,
            max: getMaxMonthFrequency(creationDates),
            inRange: {
                color: ['#d6a1ff', '#2b054a']
              },
        },
        tooltip: {},
        calendar: {
            range: '2022',
            cellSize: [16],
            monthLabel: {
                textStyle: {
                    color: '#ffffff'
                }
            },
            dayLabel: {
                textStyle: {
                    color: '#ffffff'
                }
            },
            yearLabel: {
                textStyle: {
                    color: '#ffffff'
                }
            },
        },
        series: {
            type: 'heatmap',
            coordinateSystem: 'calendar',
            data: heatmapData
        }
    };

    option && myChart.setOption(option);
}

function getContinentCoverage(countryFreqs) {
    const allCountryCodes = Object.keys(countryFreqs)
    return getContinentsCovered(allCountryCodes);
}
function fillInContinentCoverage(countryFreqs) {
    const allCountryCodes = Object.keys(countryFreqs)
    const continentCoverage = getContinentsCovered(allCountryCodes)

    document.getElementById("statBoxContinentCoverage").textContent = Math.floor(continentCoverage*100)+"%";
    removeSkeletonClasses(["statBoxContinentCoverage"])
    return
}

function removeSkeletonClasses(elementIDs) {
    elementIDs.forEach(ID => {
        document.getElementById(ID).classList.remove("skeleton");
        document.getElementById(ID).classList.remove("skeleton-text");
    })
}
function extractUniqueCountryCodesFromFriends(friends) {
    let allCountryCodes = []
    friends.forEach(friend => {
        if (friend.User.accdetails.loccountrycode != "") {
            allCountryCodes.push(friend.User.accdetails.loccountrycode)
        }
    });
    // Get rid of duplicates
    allCountryCodes = [...new Set(allCountryCodes)]
    return allCountryCodes;
}

// https://dev.to/jorik/country-code-to-flag-emoji-a21
function getFlagEmoji(countryCode) {
    const codePoints = countryCode
      .toUpperCase()
      .split('')
      .map(char =>  127397 + char.charCodeAt());
    return String.fromCodePoint(...codePoints);
}

function getContinentsCovered(countryCodes) {
    const TOTAL_CONTINENTS = 7
    let continentCoverage = 0
    let asiaMatch = false;
    let africaMatch = false;
    let australiaMatch = false;
    let europeMatch = false;
    let northAmericaMatch = false;
    let southAmericaMatch = false

    countryCodes.forEach(countryCode => {
        if (!asiaMatch) {
            if (continents["asia"].includes(countryCode.toUpperCase())) {
                continentCoverage++
                asiaMatch = true
            }
        }
        if (!africaMatch) {
            if (continents["africa"].includes(countryCode.toUpperCase())) {
                continentCoverage++
                africaMatch = true
            }
        }
        if (!australiaMatch) {
            if (continents["australia"].includes(countryCode.toUpperCase())) {
                continentCoverage++
                australiaMatch = true
            }
        }
        if (!europeMatch) {
            if (continents["europe"].includes(countryCode.toUpperCase())) {
                continentCoverage++
                europeMatch = true
            }
        }
        if (!northAmericaMatch) {
            if (continents["north america"].includes(countryCode.toUpperCase())) {
                continentCoverage++
                northAmericaMatch = true
            }
        }
        if (!southAmericaMatch) {
            if (continents["south america"].includes(countryCode.toUpperCase())) {
                continentCoverage++
                southAmericaMatch = true
            }
        }
    })
    return continentCoverage / TOTAL_CONTINENTS
}

function fillInMonthStatBoxes(creationDates) {
    let userCreationMonthFrequencies = {}
    creationDates.forEach(date => {
        userCreationMonthFrequencies[date.getMonth()] = userCreationMonthFrequencies[date.getMonth()] ? userCreationMonthFrequencies[date.getMonth()] + 1 : 1;
    })
    const sortedMonthFrequencies = Object.entries(userCreationMonthFrequencies).sort((a, b) => { return a[1] < b[1]})

    document.getElementById("statBoxMostPopularMonth").textContent = intToMonth(Object.values(sortedMonthFrequencies)[0][0])
    document.getElementById("statBoxLeastPopularMonth").textContent = intToMonth(Object.values(sortedMonthFrequencies)[sortedMonthFrequencies.length-1][0])

    removeSkeletonClasses(["statBoxMostPopularMonth", "statBoxLeastPopularMonth"])
}

// https://worldpopulationreview.com/country-rankings/dictatorship-countries  
function ruledByDictatorCountries(countries) {
    let dictatorRuledCountryCount = 0
    const dictatorRuledCountries = [
        "AF", "AL", "AO", "AZ", "BH", "BD", "BY", "BN", "BI", "KH",
        "CM", "CF", "TD", "CN", "CU", "DJ", "CD", "EG", "GQ", "ER", 
        "SZ", "ET", "GA", "IR", "IQ", "KZ", "LA", "LY", "MM", "NI",
        "KP", "OM", "QA", "CD", "RU", "RW", "SA", "SO", "SD", "SY",
        "SS", "TJ", "TR", "TM", "UG", "AE", "UZ", "VE", "VN", "EH",
        "YE"
    ]
    countries.forEach(countryCode => {
        if (dictatorRuledCountries.includes(countryCode.toUpperCase())) {
            dictatorRuledCountryCount++;
        }
    })
    return dictatorRuledCountryCount;
}

function intToMonth(month) {
    let monthName = ""
    switch (parseInt(month)) {
        case 0:
            monthName = "January"
            break
        case 1:
            monthName = "Febuary"
            break
        case 2:
            monthName = "March"
            break
        case 3:
            monthName = "April"
            break
        case 4:
            monthName = "May"
            break
        case 5:
            monthName = "June"
            break
        case 6:
            monthName = "July"
            break
        case 7:
            monthName = "August"
            break
        case 8:
            monthName = "September"
            break
        case 9:
            monthName = "October"
            break
        case 10:
            monthName = "November"
            break
        case 11:
            monthName = "December"
            break
        default:
            console.error("failed to find most popular user creation month")
            monthName = "na"
    }
    return monthName
}

function fillInOldestAndNewestUserCards(graphData) {
    let allFriends = graphData.frienddetails;
    allFriends.sort((f1,f2) => { 
        return new Date(f1.User.accdetails.timecreated * 1000) - new Date(f2.User.accdetails.timecreated * 1000) 
    })

    const oldestUser = allFriends[0].User
    const newestUser = allFriends[allFriends.length-1].User

    document.getElementById("oldestUserUsername").textContent = oldestUser.accdetails.personaname;
    document.getElementById("oldestUserRealName").textContent = "idk";
    document.getElementById("oldestUserFriendCount").textContent = oldestUser.friendids.length;
    let creationDate = new Date(oldestUser.accdetails.timecreated*1000);
    let dateString = `${creationDate.getDate()} ${creationDate.toLocaleString('default', { month: 'long' })} ${creationDate.getFullYear()}`;
    let timeSinceString = `(${timezSince(creationDate)} ago)`
    document.getElementById("oldestUserCreationDate").textContent = `${dateString} ${timeSinceString}`;
    document.getElementById("oldestUserSteamID").textContent = oldestUser.accdetails.steamid;
    document.getElementById("oldestUserAvatar").src = oldestUser.accdetails.avatar.split(".jpg").join("") + "_full.jpg";

    document.getElementById("newestUserUsername").textContent = newestUser.accdetails.personaname;
    document.getElementById("newestUserRealName").textContent = "idk";
    document.getElementById("newestUserFriendCount").textContent = newestUser.friendids.length;
    creationDate = new Date(newestUser.accdetails.timecreated*1000);
    dateString = `${creationDate.getDate()} ${creationDate.toLocaleString('default', { month: 'long' })} ${creationDate.getFullYear()}`;
    timeSinceString = `(${timezSince(creationDate)} ago)`
    document.getElementById("newestUserCreationDate").textContent = `${dateString} ${timeSinceString}`;
    document.getElementById("newestUserSteamID").textContent = newestUser.accdetails.steamid;
    document.getElementById("newestUserAvatar").src = newestUser.accdetails.avatar.split(".jpg").join("") + "_full.jpg";

    removeSkeletonClasses(["oldestUserUsername", "oldestUserRealName", 
    "oldestUserFriendCount", "oldestUserCreationDate", "oldestUserSteamID", "oldestUserAvatar",
    "newestUserUsername", "newestUserRealName", "newestUserFriendCount", "newestUserCreationDate",
    "newestUserSteamID", "newestUserAvatar"])
}

function getTopTenCountries(countriesFreq) {
    let countryNames = []
    const sortedCountriesFreq = Object.entries(countriesFreq).sort((a,b) => b[1]-a[1])
    for (let i = 0; i < sortedCountriesFreq.length; i++) {
        countryNames.push(countryCodeToName(sortedCountriesFreq[i][0].toUpperCase()))
    }
    if (countryNames.length >= 10) {
        return countryNames.slice(0, 10)
    }
    return countryNames;
}

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

const continents = {
    "asia": [
        "CN", "IN", "ID", "PK", "BD", "JP", "PH", "VN", "TR", "IR", "TH",
        "IR", "MM", "KR", "IQ", "AF", "SA", "UZ", "MY", "YE", "NP", "TW",
        "LK", "KZ", "SY", "KH", "JO", "AZ", "AE", "TJ", "IL", "HK", "LA",
        "LB", "KG", "TM", "SG", "OM", "PS", "KW", "GE", "MN", "AM", "QA",
        "BH", "TL", "CY", "BT", "MO", "MV", "BN"
    ],
    "africa": [
        "NG", "ET", "EG", "CD", "CG", "TZ", "SA", "KE", "UG", "DZ", "SD",
        "MA", "AO", "MZ", "GH", "MG", "CM", "CI", "NE", "BF", "ML", "MW",
        "ZM", "SN", "TD", "SO", "ZW", "GN", "RW", "BJ", "BI", "TN", "TG",
        "SL", "LY", "CG", "LR", "CF", "MR", "ER", "NA", "GM", "BW", "GA",
        "LS", "GW", "GQ", "MU", "DJ", "RE", "KM", "EH", "YT", "ST", "SC",
        "SH"
    ],
    "europe": [
        "RU", "DE", "GB", "FR", "IT", "ES", "UA", "PL", "RO", "NL", "BE", "CZ",
        "GR", "PT", "SE", "HU", "BY", "AT", "RS", "CH", "BG", "DK", "FI", "SK",
        "NO", "HR", "IE", "MD", "BA", "AL", "LT", "MK", "SI", "LV", "EE", "ME",
        "LU", "MT", "IS", "AD", "FO", "MC", "LI", "SM", "GI", "VA"
    ],
    "north america": [
        "US", "MX", "CA", "GT", "HT", "CU", "DO", "HN", "NI", "SV", "CR", "PA",
        "JM", "PR", "TT", "GP", "BZ", "BS", "MQ", "BB", "LC", "GD", "VC", "AW",
        "VI", "AG", "DM", "KY", "BM", "GL", "KN", "MF", "VG", "AN", "AI", "BL",
        "PM", "MS"
    ],
    "south america": [
        "BR", "CO", "AR", "PE", "VE", "CL", "EC", "BO", "PY", "UY", "SR", "GF",
        "FK"
    ],
    "australia": [
        "AU", "PG", "NZ", "FJ", "SB", "FM", "VU", "NC", "PF", "WS", "GU", "KI",
        "TO", "MH", "MP", "AS", "PW", "CK", "TB", "WF", "NR", "NU", "TK"
    ],
    "antarctica": [

    ],
}
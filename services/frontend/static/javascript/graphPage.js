import { countUpElement } from '/static/javascript/countUpScript.js';
import { setUserCardDetails } from '/static/javascript/userCard.js';
import { getHeatmapData, getMaxMonthFrequency } from './heatMapCalendarHelper.js';
import * as util from '/static/javascript/util.js';
import * as utilRequest from '/static/javascript/utilRequests.js';

const URLarr = window.location.href.split("/");
const crawlID = URLarr[URLarr.length-1];
let crawlData = {}

utilRequest.doesProcessedGraphDataExist(crawlID).then(doesExist => {
    if (doesExist === false) {
        window.location.href = "/"
    }
    utilRequest.getProcessedGraphData(crawlID).then(crawlDataObj => {
        crawlData = crawlDataObj
        console.log(crawlData.usergraphdata.userdetails.User)
        setUserCardDetails(crawlData.usergraphdata.userdetails.User);
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
        fillInUserAndNetworkFavoriteGameStatBoxes(crawlDataObj.usergraphdata)
        let usersLeaderboard = getMostHoursPlayedStats(crawlDataObj.usergraphdata)
        fillInHoursPlayedLeaderboard(usersLeaderboard)
        initNetWorkMostHoursPlayedBarChart(usersLeaderboard)

        // Friend network stats
        userCreatedGraph(crawlDataObj.usergraphdata)
        userCreatedMonthChart(crawlDataObj.usergraphdata)
        fillInOldestAndNewestUserCards(crawlDataObj.usergraphdata)
        initAndRenderAccountAgeVsFriendCountChart(crawlDataObj.usergraphdata)

        initGamerScore(crawlDataObj.usergraphdata)
        initLinkForInteractiveGraphPage()

        var myChart = echarts.init(document.getElementById('graphContainer'));
        const graph = getDataInGraphFormat(crawlDataObj.usergraphdata, countryFrequencies)
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
                }),
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

function getDataInGraphFormat(gData, countryFrequencies) {
    const topTenCountryNames = getTopTenCountries(countryFrequencies);
    // TODO change to top 10 frequency countries instead
    let countryCategories = []
    topTenCountryNames.forEach(countryName => {
        countryCategories.push({ "name": countryName })
    })
    countryCategories.push({"name":"Other"})
    let nodes = []
    let links = []

    let usersCountryName = util.countryCodeToName(gData.userdetails.User.accdetails.loccountrycode.toUpperCase())
    let usersCountryCategory = topTenCountryNames.includes(usersCountryName) ? usersCountryName : 'Other';
    nodes.push({
        "id": gData.userdetails.User.accdetails.steamid,
        "name": gData.userdetails.User.accdetails.personaname,
        "value": gData.userdetails.User.accdetails.profileurl,
        "category": usersCountryCategory,
        "symbol": `image://${gData.userdetails.User.accdetails.avatar}`
    })

    gData.frienddetails.forEach((friend) => {
        let usersCountryName = util.countryCodeToName(friend.User.accdetails.loccountrycode.toUpperCase())
        let usersCountryCategory = topTenCountryNames.includes(usersCountryName) ? usersCountryName : 'Other';
        nodes.push({
            "id": friend.User.accdetails.steamid,
            "name": friend.User.accdetails.personaname,
            "value": friend.User.accdetails.profileurl,
            "category": usersCountryCategory,
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
        "categories": countryCategories
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
        percentageOfFriendsWithLessHoursPlayed = Math.floor((friendsWithLessHoursPlayed / graphData.frienddetails.length) * 100)
    }

    countUpElement("statBoxHoursAcrossLibrary", totalHoursPlayedForUser)
    countUpElement("statBoxEntireDaysOfPlaytime", entireDaysOfPlaytime)
    countUpElement("statBoxFriendsWithLessHoursPlayed", percentageOfFriendsWithLessHoursPlayed, {suffix: "%"})
    countUpElement("statBoxMinWageEarned", minWageEarnedForGaming, {prefix: "€"})

    util.removeSkeletonClasses(["statBoxHoursAcrossLibrary", "statBoxEntireDaysOfPlaytime",
            "statBoxMinWageEarned", "statBoxFriendsWithLessHoursPlayed"])
}

function fillInUserAndNetworkFavoriteGameStatBoxes(graphData) {
    const steamGameInfoAPI = "/getgamedetails"

    const usersFavoriteGame = graphData.userdetails.User.gamesowned[0];
    if (usersFavoriteGame != undefined) {
        console.log(`${steamGameInfoAPI}/${usersFavoriteGame.appid}`)
        fetch(`${steamGameInfoAPI}/${usersFavoriteGame.appid}`)
        .then(res => res.json())
        .then(res => {
            document.getElementById("userFavoriteGameName").textContent = res[usersFavoriteGame.appid].data.name;
            document.getElementById("userFavoriteGameImage").src = res[usersFavoriteGame.appid].data.header_image;
            document.getElementById("userFavoriteGameShopPage").href = res[usersFavoriteGame.appid].data.website
            
            const playtimeInHours = Math.floor(usersFavoriteGame.playtime_forever/60)
            countUpElement('statBoxUsersFavoriteGameHoursPlayed', playtimeInHours)
            const gameCost = res[usersFavoriteGame.appid].data.price_overview;
            let friendsWhoPlayUsersFavoriteGame = 0
            graphData.frienddetails.forEach(user => {
                const friend = user.User;
                friend.gamesowned.forEach(game => {
                    if (game.appid === usersFavoriteGame.appid) {
                        friendsWhoPlayUsersFavoriteGame++
                    }
                })
            })
            document.getElementById("userFavoriteGameXFriendsAlsoPlay").textContent = `${friendsWhoPlayUsersFavoriteGame} friends also play this`;

            let costPerHour = 0;
            if (gameCost) {
                console.log(gameCost)
                console.log((gameCost.initial/100))
                console.log(playtimeInHours)
                console.log(((gameCost.initial/100)/playtimeInHours))
                costPerHour = ((gameCost.initial/100)/playtimeInHours).toFixed(2);
            }
            document.getElementById('statBoxUsersFavoriteGamesCostPerHour').textContent = `€${costPerHour}`

            util.removeSkeletonClasses(["userFavoriteGameName", "statBoxUsersFavoriteGameHoursPlayed", 
                "statBoxUsersFavoriteGamesCostPerHour", "userFavoriteGameXFriendsAlsoPlay",
                "userFavoriteGameImage"])
        })
    }

    const networksFavoriteGame = graphData.topgamedetails[0]
    if (networksFavoriteGame != undefined) {
        fetch(`${steamGameInfoAPI}/${networksFavoriteGame.appid}`)
        .then(res => res.json())
        .then(res => {
            document.getElementById("networkFavoriteGameName").textContent = res[networksFavoriteGame.appid].data.name;
            document.getElementById("networkFavoriteGameImage").src = res[networksFavoriteGame.appid].data.header_image;
            document.getElementById("networkFavoriteGameShopPage").href = res[networksFavoriteGame.appid].data.website
            
            let hoursPlayedByNetwork = []
            graphData.frienddetails.forEach(user => {
                const friend = user.User;
                friend.gamesowned.forEach(game => {
                    if (game.appid === networksFavoriteGame.appid) {
                        hoursPlayedByNetwork.push(Math.floor(game.playtime_forever/60))
                    }
                })
            })
            const totalHoursByNetwork = hoursPlayedByNetwork.reduce((sum, hours) => sum += hours, 0);
            countUpElement('statBoxNetworksFavoriteGameHoursPlayed', totalHoursByNetwork)
            const gameCost = res[networksFavoriteGame.appid].data.price_overview;
            document.getElementById("networkFavoriteGameXFriendsAlsoPlay").textContent = `${hoursPlayedByNetwork.length} friends play this`;
            
            let costPerHour = 0;
            if (gameCost) {
                const totalCostOfGames = (gameCost.initial/100) * hoursPlayedByNetwork.length
                costPerHour = (totalCostOfGames/totalHoursByNetwork).toFixed(2);
            }
            document.getElementById('statBoxNetworksFavoriteGameCostPerHourAverage').textContent = `€${costPerHour}`

            util.removeSkeletonClasses(["networkFavoriteGameName", "statBoxNetworksFavoriteGameHoursPlayed", 
                "statBoxNetworksFavoriteGameCostPerHourAverage", "networkFavoriteGameXFriendsAlsoPlay",
                "networkFavoriteGameImage"])
        })
    }
}


function initAndRenderAccountAgeVsFriendCountChart(graphData) {
    let scatterPlotData = []
    let maxAccountAge = 0

    let highestFriendCountUser;
    let maxFriends = 0;
    graphData.frienddetails.forEach(user => {
        const friends = user.User.friendids.length;
        if (friends > maxFriends) {
            highestFriendCountUser = user;
            maxFriends = friends
        }
        const accAge = user.User.accdetails.timecreated;
        let monthsSinceCreation = util.monthsSince(accAge)
        if (monthsSinceCreation > maxAccountAge) {
            maxAccountAge = monthsSinceCreation
        }
        scatterPlotData.push([
            friends, monthsSinceCreation
        ])
    })
    highestFriendCountUser = highestFriendCountUser.User;
    document.getElementById("highestFriendCountUserUsername").textContent = highestFriendCountUser.accdetails.personaname;
    document.getElementById("highestFriendCountUserCountry").textContent = util.countryCodeToName(highestFriendCountUser.accdetails.loccountrycode) === "" ? 'unknown' : util.countryCodeToName(highestFriendCountUser.accdetails.loccountrycode);
    document.getElementById("highestFriendCountUserFriendCount").textContent = highestFriendCountUser.friendids.length;
    let creationDate = new Date(highestFriendCountUser.accdetails.timecreated*1000);
    let dateString = `${creationDate.getDate()} ${creationDate.toLocaleString('default', { month: 'long' })} ${creationDate.getFullYear()}`;
    let timeSinceString = `(${util.timezSince(creationDate)} ago)`
    document.getElementById("highestFriendCountUserCreationDate").textContent = `${dateString} ${timeSinceString}`;
    document.getElementById("highestFriendCountUserProfile").innerHTML = `<a href="${highestFriendCountUser.accdetails.profileurl}">Profile Link</a>`;
    document.getElementById("highestFriendCountUserAvatar").src = highestFriendCountUser.accdetails.avatar.split(".jpg").join("") + "_full.jpg";
   
    util.removeSkeletonClasses(["highestFriendCountUserUsername", "highestFriendCountUserCountry", 
        "highestFriendCountUserFriendCount", "highestFriendCountUserCreationDate", 
        "highestFriendCountUserCountry", "highestFriendCountUserAvatar", "highestFriendCountUserProfile"])


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

function initLinkForInteractiveGraphPage() {
    document.getElementById("interactiveGraphLink").href = `/graph/interactive?firstcrawlid=${crawlID}`
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
        return 80 + (link.source.neighbourNodes.length * 8);
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
        // limit the amount of flags that can be rendered
        if (i == 48) {
            return
        }
        const usersFromCountryFormatted = `Friends from ${util.countryCodeToName(countryCode)}:\n${getUsernamesFromCountry(friends, countryCode).join("\n")}`
        document.getElementById("allFlagsDiv").innerHTML += `
        <div class="col-1">
            <p style="font-size: 1.7rem" title="${usersFromCountryFormatted}">${util.getFlagEmoji(countryCode)}</p>
        </div>
        `;
        i++;
    });
}

function getUsernamesFromCountry(friends, countryCode) {
    let usersFromCountry = []
    friends.forEach(friend => {
        if (friend.User.accdetails.loccountrycode == countryCode) {
            usersFromCountry.push(friend.User.accdetails.personaname)
        }
    });
    return usersFromCountry;
}

function generateOverallGamerScore(graphData) {
    const totalHoursPlayed  = getHoursPlayedForUser(graphData.userdetails.User);
    const totalFriends = graphData.userdetails.User.friendids.length;

    const gamerScore = (totalHoursPlayed * 2) + totalFriends
    console.log(gamerScore)
    return gamerScore >= 5000 ? 5000 : gamerScore 
}

function getMostHoursPlayedStats(graphData) {
    let allUsers = []
    
    const mainUserObj = {
        "username": graphData.userdetails.User.accdetails.personaname,
        "profiler": graphData.userdetails.User.accdetails.avatar,
        "profileURL": graphData.userdetails.User.accdetails.profileurl,
        "hours": getHoursPlayedForUser(graphData.userdetails.User)
    }
    allUsers.push(mainUserObj)

    graphData.frienddetails.forEach(user => {
        const friend = user.User;
        allUsers.push({
            "username": friend.accdetails.personaname,
            "profiler": friend.accdetails.avatar,
            "profileURL": friend.accdetails.profileurl,
            "hours": getHoursPlayedForUser(friend)
        })
    }) 
    allUsers.sort(function(userOne, userTwo) {
        return userOne.hours < userTwo.hours ? 1 : userTwo.hours < userOne.hours ? -1 : 0
    })

    for (let i = 0; i < allUsers.length; i++) {
        const currUser = allUsers[i]
        if (currUser.avatar === graphData.userdetails.User.accdetails.avatar &&
            currUser.username === graphData.userdetails.User.accdetails.personaname) {

                let topEightUsers = allUsers.length >= 8 ? allUsers.slice(0, 8) : allUsers;
                const returnObj = {
                        "users": topEightUsers
                }
                return returnObj
        }
    }
    // The main user is not in the top 8. Include them to be displayed seperately
    let topEightUsers = allUsers.length >= 8 ? allUsers.slice(0, 8) : allUsers;
    const returnObj = {
        "users": topEightUsers,
        "mainUser": mainUserObj
    }
    return returnObj
}

function getHoursPlayedForUser(user) {
    let totalHours = 0
    user.gamesowned.forEach(game => {
        totalHours += Math.floor(game.playtime_forever/60)
    })
    return totalHours
}

function fillInHoursPlayedLeaderboard(leaderboardData) {
    let htmlContent = ``
    const backgroundColors = ['#292929', '#414141']
    let i = 0;
    leaderboardData.users.forEach(user => {
        htmlContent += `
        <div class="row justify-content-start mt-1 pb-1" style="font-size: 1.07rem; border-radius: 6px; background-color:${backgroundColors[i%backgroundColors.length]}; border-color: white;">
                    <div class="col-1 truncate pt-1 text-center">
                        ${i+1}.
                    </div>
                    <div class="col-1 text-center">
                        <a href="${user.profileURL}">
                            <img
                                src="${user.profiler}"
                                style="height: 100%; width: auto"
                            >
                        </a>
                    </div>
                    <div class="col-8 truncate pt-1">
                        ${user.username}
                    </div>
                    <div class="col-2 truncate pt-1">
                        ${user.hours}
                    </div>
                </div>`
        i++
    })
    
    if (leaderboardData.mainUser != undefined) {
        htmlContent += `
        <div class="row mt-3 mb-0 justify-content-start mt-1 pb-1" style="font-size: 1.07rem; border: 2px solid white; border-radius: 6px; background-color:${backgroundColors[i%backgroundColors.length]}; border-color: white;">
                    <div class="col-1 truncate pt-1 text-center">
                        
                    </div>
                    <div class="col-1 text-center">
                        <a href="${leaderboardData.mainUser.profileURL}">
                            <img
                                src="${leaderboardData.mainUser.profiler}"
                                style="height: 100%; width: auto"
                            >
                        </a>
                    </div>
                    <div class="col-8 truncate pt-1">
                        ${leaderboardData.mainUser.username}
                    </div>
                    <div class="col-2 truncate pt-1">
                        ${leaderboardData.mainUser.hours}
                    </div>
                </div>`
    }
    document.getElementById("hoursPlayedLeaderboard").innerHTML = htmlContent;
}

function initNetWorkMostHoursPlayedBarChart(chartData) {

    let chartDom = document.getElementById('networkMostHoursPlayedBarChart');
    let myChart = echarts.init(chartDom);
    let option;
    let app = {}

    let usernames = []
    let hoursPlayed = []

    chartData.users.forEach(user => {

        usernames.unshift(user.username)
        hoursPlayed.unshift(user.hours)
    })

    app.config = {
        rotate: 0,
        horizontalAlign: 'middle',
        align: 'left',
        distance: 0
    }
    const labelOption = {
        show: true,
        distance: app.config.distance,
        rotate: app.config.rotate,
        verticalAlign: app.config.verticalAlign,
        align: app.config.align,
        position: app.config.position,
        formatter: '{b}',
    }

    option = {
        title: {
            text: 'Most hours played chart',
            textStyle: {
                color: '#ffffff'
            }
        },
        tooltip: {
            trigger: 'axis',
            axisPointer: {
                type: 'shadow'
            }
        },
        legend: {},
        grid: {
            left: '3%',
            right: '4%',
            bottom: '3%',
            containLabel: true
        },
        xAxis: {
            type: 'value',
            boundaryGap: [0, 0.01],
            axisLine: {
                lineStyle: {
                    color: '#ffffff'
                }
            }
        },
        yAxis: {
            type: 'category',
            data: usernames,
            axisLabel: {
                show: false
            }
        },
        series: [{
                type: 'bar',
                data: hoursPlayed,
                label: labelOption,
                barWidth: '87%',
                barCategoryGap: '115%'
            }]
    };

    option && myChart.setOption(option);
}

function fillInTopStatBoxes(graphData, countryFreqs) {
    const UNCountries = 195;
    let uniqueCountryCodes = extractUniqueCountryCodesFromFriends(graphData.usergraphdata.frienddetails)

    countUpElement('statBoxFriendCount', graphData.usergraphdata.userdetails.User.friendids.length)
    countUpElement('statBoxUniqueCountries', uniqueCountryCodes.length)
    countUpElement('statBoxGlobalCoverage', Math.floor((uniqueCountryCodes.length/UNCountries)*100), {suffix: "%"})
    countUpElement('statBoxDictatorships', ruledByDictatorCountries(uniqueCountryCodes))

    util.removeSkeletonClasses(["statBoxFriendCount", "statBoxUniqueCountries", 
            "statBoxGlobalCoverage", "statBoxDictatorships"])
}

function fillInFromYourCountryStatBox(graphDataObj, countryFreq) {
    let alsoFromCountry = 0;
    const usersCountry = graphDataObj.usergraphdata.userdetails.User.accdetails.loccountrycode;
    if (usersCountry === undefined) {
        document.getElementById("statBoxAlsoFromYourCountry").textContent = alsoFromCountry;
        util.removeSkeletonClasses(["statBoxAlsoFromYourCountry"])
        return
    }
    const otherFriendsInUsersCountry = countryFreq[usersCountry.toLowerCase()];
    if (!isNaN(otherFriendsInUsersCountry)) {
        document.getElementById("statBoxAlsoFromYourCountry").textContent = otherFriendsInUsersCountry;
        util.removeSkeletonClasses(["statBoxAlsoFromYourCountry"])
        return
    }
    document.getElementById("statBoxAlsoFromYourCountry").textContent = alsoFromCountry;
    util.removeSkeletonClasses(["statBoxAlsoFromYourCountry"])
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

function fillInContinentCoverage(countryFreqs) {
    const allCountryCodes = Object.keys(countryFreqs)
    const continentCoverage = getContinentsCovered(allCountryCodes)

    document.getElementById("statBoxContinentCoverage").textContent = Math.floor(continentCoverage*100)+"%";
    util.removeSkeletonClasses(["statBoxContinentCoverage"])
    return
}

function initGamerScore(mainUser) {
    let chartDom = document.getElementById('gamerScore');
    let myChart = echarts.init(chartDom);
    let option;

    option = {
    series: [
        {
        type: 'gauge',
        startAngle: 180,
        endAngle: 0,
        min: 0,
        max: 1,
        splitNumber: 8,
        axisLine: {
            lineStyle: {
            width: 8,
            color: [
                [0.25, '#FF6E76'],
                [0.5, '#FDDD60'],
                [0.75, '#58D9F9'],
                [1, '#7CFFB2']
            ]
            }
        },
        pointer: {
            icon: 'path://M12.8,0.7l12,40.1H0.7L12.8,0.7z',
            length: '12%',
            width: 20,
            offsetCenter: [0, '-60%'],
            itemStyle: {
            color: 'auto'
            }
        },
        axisTick: {
            length: 12,
            lineStyle: {
            color: 'auto',
            width: 2
            }
        },
        splitLine: {
            length: 20,
            lineStyle: {
            color: 'auto',
            width: 5
            }
        },
        axisLabel: {
            color: '#ffffff',
            fontSize: 30,
            distance: -60,
            formatter: function (value) {
                if (value === 5000) {
                    return 'S';
                } else if (value === 4000) {
                    return 'B';
                } else if (value === 3000) {
                    return 'B';
                } else if (value === 2000) {
                    return 'C';
                } else if (value === 1000) {
                    return 'D';
                }
                return '';
            }
        },
        title: {
            offsetCenter: [0, '-20%'],
            fontSize: 110
        },
        detail: {
            fontSize: 90,
            textStyle: {
                color: '#ffffff'
            },
            offsetCenter: [0, '0%'],
            valueAnimation: true,
            formatter: function (value) {
                return Math.round(value * 100);
            },
            color: 'auto'
        },
        data: [
            {
            value: generateOverallGamerScore(mainUser)/5000,
            }
        ]
        }
    ]
    };
    option && myChart.setOption(option);
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
            if (util.continents["asia"].includes(countryCode.toUpperCase())) {
                continentCoverage++
                asiaMatch = true
            }
        }
        if (!africaMatch) {
            if (util.continents["africa"].includes(countryCode.toUpperCase())) {
                continentCoverage++
                africaMatch = true
            }
        }
        if (!australiaMatch) {
            if (util.continents["australia"].includes(countryCode.toUpperCase())) {
                continentCoverage++
                australiaMatch = true
            }
        }
        if (!europeMatch) {
            if (util.continents["europe"].includes(countryCode.toUpperCase())) {
                continentCoverage++
                europeMatch = true
            }
        }
        if (!northAmericaMatch) {
            if (util.continents["north america"].includes(countryCode.toUpperCase())) {
                continentCoverage++
                northAmericaMatch = true
            }
        }
        if (!southAmericaMatch) {
            if (util.continents["south america"].includes(countryCode.toUpperCase())) {
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

    document.getElementById("statBoxMostPopularMonth").textContent = util.intToMonth(Object.values(sortedMonthFrequencies)[0][0])
    document.getElementById("statBoxLeastPopularMonth").textContent = util.intToMonth(Object.values(sortedMonthFrequencies)[sortedMonthFrequencies.length-1][0])

    util.removeSkeletonClasses(["statBoxMostPopularMonth", "statBoxLeastPopularMonth"])
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

function fillInOldestAndNewestUserCards(graphData) {
    let allFriends = graphData.frienddetails;
    allFriends.sort((f1,f2) => { 
        return new Date(f1.User.accdetails.timecreated * 1000) - new Date(f2.User.accdetails.timecreated * 1000) 
    })

    const oldestUser = allFriends[0].User
    const newestUser = allFriends[allFriends.length-1].User

    document.getElementById("oldestUserUsername").textContent = oldestUser.accdetails.personaname;
    document.getElementById("oldestUserCountry").textContent = util.countryCodeToName(oldestUser.accdetails.loccountrycode) === "" ? 'unknown' : util.countryCodeToName(oldestUser.accdetails.loccountrycode);
    document.getElementById("oldestUserFriendCount").textContent = oldestUser.friendids.length;
    let creationDate = new Date(oldestUser.accdetails.timecreated*1000);
    let dateString = `${creationDate.getDate()} ${creationDate.toLocaleString('default', { month: 'long' })} ${creationDate.getFullYear()}`;
    let timeSinceString = `(${util.timezSince(creationDate)} ago)`
    document.getElementById("oldestUserCreationDate").textContent = `${dateString} ${timeSinceString}`;
    document.getElementById("oldestUserProfile").innerHTML = `<a href="${oldestUser.accdetails.profileurl}">Profile Link</a>`;
    document.getElementById("oldestUserAvatar").src = oldestUser.accdetails.avatar.split(".jpg").join("") + "_full.jpg";

    document.getElementById("newestUserUsername").textContent = newestUser.accdetails.personaname;
    document.getElementById("newestUserCountry").textContent = util.countryCodeToName(newestUser.accdetails.loccountrycode) === "" ? 'unknown' : util.countryCodeToName(newestUser.accdetails.loccountrycode);
    document.getElementById("newestUserFriendCount").textContent = newestUser.friendids.length;
    creationDate = new Date(newestUser.accdetails.timecreated*1000);
    dateString = `${creationDate.getDate()} ${creationDate.toLocaleString('default', { month: 'long' })} ${creationDate.getFullYear()}`;
    timeSinceString = `(${util.timezSince(creationDate)} ago)`
    document.getElementById("newestUserCreationDate").textContent = `${dateString} ${timeSinceString}`;
    document.getElementById("newestUserProfile").innerHTML = `<a href="${newestUser.accdetails.profileurl}">Profile Link</a>`;
    document.getElementById("newestUserAvatar").src = newestUser.accdetails.avatar.split(".jpg").join("") + "_full.jpg";

    util.removeSkeletonClasses(["oldestUserUsername", "oldestUserCountry", 
    "oldestUserFriendCount", "oldestUserCreationDate", "oldestUserProfile", "oldestUserAvatar",
    "newestUserUsername", "newestUserCountry", "newestUserFriendCount", "newestUserCreationDate",
    "newestUserProfile", "newestUserAvatar"])
}

function getTopTenCountries(countriesFreq) {
    let countryNames = []
    const sortedCountriesFreq = Object.entries(countriesFreq).sort((a,b) => b[1]-a[1])
    for (let i = 0; i < sortedCountriesFreq.length; i++) {
        countryNames.push(util.countryCodeToName(sortedCountriesFreq[i][0].toUpperCase()))
    }
    if (countryNames.length >= 10) {
        return countryNames.slice(0, 10)
    }
    return countryNames;
}
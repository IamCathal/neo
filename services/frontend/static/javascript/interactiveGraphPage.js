const crawlIDs = getCrawlIDs()
console.log("Eeeeeeeeeeeeee")

if (crawlIDs.length == 2) {
    console.log("two IDS: " + crawlIDs)
    // Get graph data for both users and merge

    const firstUserProcessedGraphData = getProcessedGraphData(crawlIDs[0])
    const secondUserProcessedGraphData = getProcessedGraphData(crawlIDs[0])

    Promise.all([firstUserProcessedGraphData, secondUserProcessedGraphData]).then(graphDatas => {
        console.log("now I have both lots of graph data")
        console.log(graphDatas[0])
        console.log(graphDatas[1])

        combinedGraphData = {
            "firstuser": graphDatas[0].usergraphdata.userdetails.User,
            "seconduser": graphDatas[1].usergraphdata.userdetails.User,
            "allfriends": combineNetworks(graphDatas[0].usergraphdata, graphDatas[1].usergraphdata)
        }
        console.log(combinedGraphData)

        getShortestDistance(crawlIDs).then(shortestDistanceInfo => {
            console.log(shortestDistanceInfo)
            combinedGraphData["shortestdistance"] = shortestDistanceInfo.shortestdistance
            initThreeJSGraphForTwoUsersCombined(shortestDistanceInfo)
        }, err => {
            console.error(`error getting shortest distance info: ${err}`)
        })
    }, errs => {
        console.error(`error(s) retrieving graph datas: ${errs}`);
    })
    
} 
if (crawlIDs.length == 1) {
    console.log("one id: "+ crawlIDs)
    doesProcessedGraphDataExistz(crawlIDs[0]).then(doesExist => {
        if (doesExist === false) {
            window.location.href = "/"
        }
        getProcessedGraphData(crawlIDs[0]).then(crawlDataObj => {
            initThreeJSGraph(crawlDataObj.usergraphdata)
        }, err => {
            console.error(`error retrieving processed graph data: ${err}`)
        })
    }, err => {
        console.error(`error calling does processed graphdata exist: ${err}`)
    })
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

function initThreeJSGraphForTwoUsersCombined(crawlData) {
    const shortestDistanceIDs = crawlData.shortestdistance.map(n => n.accdetails.steamid)
    let seenNodes = new Map()
    let nodes = []
    let links = []

    // Add all the required nodes
    nodes.push({
        "id": crawlData.firstuser.accdetails.steamid, 
        "username": crawlData.firstuser.accdetails.personaname,
        "avatar":crawlData.firstuser.accdetails.avatar
    })
    seenNodes.set(crawlData.firstuser.accdetails.steamid, true)
    nodes.push({
        "id": crawlData.seconduser.accdetails.steamid, 
        "username": crawlData.seconduser.accdetails.personaname,
        "avatar":crawlData.seconduser.accdetails.avatar
    })
    seenNodes.set(crawlData.seconduser.accdetails.steamid, true)

    crawlData.allfriends.forEach(friend => {
        nodes.push({
            "id":friend.accdetails.steamid, 
            "username": friend.accdetails.personaname,
            "avatar":friend.accdetails.avatar
        })
        seenNodes.set(friend.accdetails.steamid, true)
    })

    // Add all the required edges
    crawlData.firstuser.friendids.forEach(ID => {
        links.push({
            "source": crawlData.firstuser.accdetails.steamid,
            "target": ID,
            "color": linkInShortestPath(crawlData.firstuser.accdetails.steamid, ID, shortestDistanceIDs) ? 'green' : 'white'
        })
    })
    crawlData.seconduser.friendids.forEach(ID => {
        links.push({
            "source": crawlData.seconduser.accdetails.steamid,
            "target": ID,
            "color": linkInShortestPath(crawlData.seconduser.accdetails.steamid, ID, shortestDistanceIDs) ? 'green' : 'white'
        })
    })
    crawlData.allfriends.forEach(friend => {
        friend.friendids.forEach(ID => {
            if (seenNodes.has(ID)) {
                links.push({
                    "source": friend.accdetails.steamid,
                    "target": ID,
                    "color": linkInShortestPath(friend.accdetails.steamid, ID, shortestDistanceIDs) ? 'green' : 'white'
                })
            }
        })
    })

    console.log(crawlData)
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
    console.log(threeJSGraphData)

    const threeJSGraphDiv = document.getElementById('3d-graph');
    let hoveredNode = null;
    let highlightedNodes = new Set()
    let highlightedLinks = new Set()
    const g = ForceGraph3D()(threeJSGraphDiv)
        .graphData(threeJSGraphData)
        // .linkColor(link => {
        //     if (highlightedLinks.has(link)) {
        //         return '#00ffcf';
        //     }
        //     if (shortestDistanceIDs.includes(link.source) && shortestDistanceIDs.includes(link.target)) {
        //         return 'green'
        //     } else {
        //         return 'white'
        //     }
        // })
        // .linkAutoColorBy(link => {
        //     if (highlightedLinks.has(link)) {
        //         return '#00ffcf';
        //     }
        //     if (shortestDistanceIDs.includes(link.source) && shortestDistanceIDs.includes(link.target)) {
        //         return 'green'
        //     } else {
        //         return 'white'
        //     }
        // })
        .nodeThreeObject(({ avatar }) => {
            const imgTexture = new THREE.TextureLoader().load(avatar);
            const material = new THREE.SpriteMaterial({ map: imgTexture });
            const sprite = new THREE.Sprite(material);
            sprite.scale.set(19, 19);
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
        .linkWidth(link => {
            if (highlightedLinks.has(link)) { 
                return 4
            }
            if (shortestDistanceIDs.includes(link.source) && shortestDistanceIDs.includes(link.target)) {
                return 8;
            } else {
                return 1;
            }
        });
        // .linkWidth(link => highlightedLinks.has(link) ? 4 : 1)
        // .linkColor(link => highlightedLinks.has(link) ? 'green' : 'white')
        // .linkDirectionalParticles(link => highlightedLinks.has(link) ? 8 : 0)
        // .linkDirectionalParticleWidth(3)
        // .linkDirectionalParticleColor(() => 'green')
        // .onNodeHover(node => {
        //     if ((!node && !highlightedNodes.size) || (node && hoveredNode === node)) {
        //         return;
        //     }

        //     highlightedLinks.clear()
        //     highlightedNodes.clear()
        //     if (node != undefined && node != false) {
        //         highlightedNodes.add(node)
        //         node.neighbourNodes.forEach(neighourNode => {
        //             highlightedNodes.add(neighourNode);
        //         });
        //         node.neighbourLinks.forEach(neighbourLink => {
        //             highlightedLinks.add(neighbourLink)
        //         })
        //     }

        //     hoveredNode = node || null;

        //     g.nodeColor(g.nodeColor())
        //         .linkWidth(g.linkWidth())
        //         .linkDirectionalParticles(g.linkDirectionalParticles());
        // });

    const linkForce = g
    .d3Force("link")
    .distance(link => {
        return 80 + (link.source.neighbourNodes.length * 8);
    });
}

function combineNetworks(firstUser, secondUser) {
    let allUsers = []
    let seenUsers = new Map()
    seenUsers.set(firstUser.userdetails.User.accdetails.steamid, true)
    seenUsers.set(secondUser.userdetails.User.accdetails.steamid, true)

    firstUser.frienddetails.forEach(user => {
        const friend = user.User
        seenUsers.set(friend.accdetails.steamid, true)
        allUsers.push(friend.accdetails)
    })
    secondUser.frienddetails.forEach(user => {
        const friend = user.User
        seenUsers.set(friend.accdetails.steamid, true)
        allUsers.push(friend.accdetails)
    })
    return allUsers
}

function linkInShortestPath(src, dest, shortestPath) {
    if (shortestPath.includes(src) && shortestPath.includes(dest)) {
        return true
    }
    return false
}
// COMMON
function getCrawlIDs() {
    const queryParams = new URLSearchParams(window.location.search);
    const firstCrawlID = queryParams.get("firstcrawlid")
    const secondCrawlID = queryParams.get("secondcrawlid")

    if (secondCrawlID != undefined) {
        console.log(`return crawlids: ${[firstCrawlID, secondCrawlID]}`)
        return [firstCrawlID, secondCrawlID]
    } else {
        console.log(`return crawlid: ${[firstCrawlID]}`)
        return [firstCrawlID]
    }
}

// COMMON 
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

function getProcessedGraphData(crawlID) {
    return new Promise((resolve, reject) => {
        fetch(`http://localhost:2590/api/getprocessedgraphdata/${crawlID}`, {
            method: "POST",
            headers: {
                'Accept-Encoding': 'gzip'
            }
        }).then(res => res.json())
        .then(data => {
            resolve(data)
        }).catch(err => {
            reject(err)
        })
    });
}
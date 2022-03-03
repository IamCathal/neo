const URLarr = window.location.href.split("/");
const crawlID = URLarr[URLarr.length-2];

console.log(crawlID)

doesProcessedGraphDataExistz(crawlID).then(doesExist => {
    if (doesExist === false) {
        window.location.href = "/"
    }
    getProcessedGraphData(crawlID).then(crawlDataObj => {
        initThreeJSGraph(crawlDataObj.usergraphdata)
    }, err => {
        console.error(`error retrieving processed graph data: ${err}`)
    })
}, err => {
    console.error(`error calling does processed graphdata exist: ${err}`)
})


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
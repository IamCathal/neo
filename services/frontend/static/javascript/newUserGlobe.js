const config = {
    speed: 0.008,
    vertTilt: -30,
    horizTilt: 20
}
const width = 958;
const height = 585;

let newUserLocations = [];
const svg = d3.select('svg')
    .attr('width', width).attr('height', height);
const markerGroup = svg.append('g');
const projection = d3.geoOrthographic();
const initialScale = projection.scale();
const path = d3.geoPath().projection(projection);
const center = [width/2, height/2];

let frameCount = 1;

drawGlobe();    
drawGlobeConstructionLines();
enableGlobeRotation();    

function drawGlobe() {  
    d3.queue()
        .defer(d3.json, 'https://cdn.jsdelivr.net/npm/world-atlas@2.0.2/countries-110m.json')             
        .await((error, worldData) => {
            if (error != undefined) {
                console.error(`error getting countries json: ${error}`)
            }
            svg.selectAll(".segment")
                .data(topojson.feature(worldData, worldData.objects.countries).features)
                .enter().append("path")
                .attr("class", "segment")
                .attr("d", path)
                .style("stroke", "#ababa9")
                .style("stroke-width", "0px")
                .style("fill", (i, j) => '#4e9455')
                drawNewUserLocations();                   
        });
}

function drawGlobeConstructionLines() {
    // Number of construction lines on x and y axis
    const graticule = d3.geoGraticule()
        .step([8, 8]);
    svg.append("path")
        .datum(graticule)
        .attr("class", "graticule")
        .attr("d", path)
        .style("opacity", "0.3")
        .style("stroke", "#acacac")
        .style("fill", "#b7e7f4");
}

function enableGlobeRotation() {
    d3.timer(function (elapsed) {
        // config.horizTilt += 0.02
        // config.vertTilt += 0.07
        projection.rotate([config.speed * elapsed, config.vertTilt, config.horizTilt]);
        svg.selectAll("path").attr("d", path);
        drawNewUserLocations();
    });
}        

function drawNewUserLocations() {
    const blipSize = frameCount % 100;
    const markers = markerGroup.selectAll('circle')
        .data(newUserLocations);
    markers
        .enter()
        .append('circle')
        .merge(markers)
        .attr('cy', point => projection([point.longitude, point.latitude])[1])
        .attr('cx', point => projection([point.longitude, point.latitude])[0])
        .attr("stroke", point => {
            const coord = [point.longitude, point.latitude];
            let sphereDistance = d3.geoDistance(coord, projection.invert(center));
            return sphereDistance > 1.58 ? '' : 'black';
        })
        .attr("stroke-width", "2")
        .attr("fill", "none")
        .attr('r', (blipSize * 0.025));
    frameCount += 2

    markerGroup.each(function () {
        this.parentNode.appendChild(this);
    });
}
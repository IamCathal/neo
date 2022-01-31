export function getHeatmapData(userCreationDates) {
    let userCreationMonthFrequencies = {}
    let heatMapData = []

    userCreationDates.forEach(date => {
        userCreationMonthFrequencies[date.getMonth()] = userCreationMonthFrequencies[date.getMonth()] ? userCreationMonthFrequencies[date.getMonth()] + 1 : 1;
    })
    const monthLengthsInDays = [31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31]
    let i = 0;
    monthLengthsInDays.forEach(monthLength => {
        for (let k = 0; k <= monthLength; k++) {
            let currTime = echarts.number.parseDate(`2022-${i+1}-${k}`)
            heatMapData.push([
                echarts.format.formatTime('yyyy-MM-dd', currTime),
                Object.values(userCreationMonthFrequencies)[i]
            ]);
        }
        i++;
    })
    return heatMapData;
}

export function getMaxMonthFrequency(userCreationDates) {
    let userCreationMonthFrequencies = {}
    userCreationDates.forEach(date => {
        userCreationMonthFrequencies[date.getMonth()] = userCreationMonthFrequencies[date.getMonth()] ? userCreationMonthFrequencies[date.getMonth()] + 1 : 1;
    })
    return Object.values(userCreationMonthFrequencies).sort((a,b) => { return b-a})[0]
}
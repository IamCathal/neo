import { CountUp } from 'https://cdnjs.cloudflare.com/ajax/libs/countup.js/2.0.7/countUp.js';

export function countUpElement(elementID, value, options) {
    const countUpElem = new CountUp(elementID, value, options);
    if (countUpElem.error) {
        console.error(`Error initialising countUp for element '${elementID}': ${countUpElem.error}`);
    } else {
        countUpElem.start()
    }
}
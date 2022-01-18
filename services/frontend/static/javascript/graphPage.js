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

// COMMON
function setUserCardDetailZ(userObj) {
    document.getElementById("userUsername").textContent = userObj.User.accdetails.personaname;
    document.getElementById("userRealName").textContent = "idk";
    document.getElementById("userFriendCount").textContent = userObj.User.friendids.length;
    
    const creationDate = new Date(userObj.User.accdetails.timecreated*1000);
    const dateString = `${creationDate.getDate()} ${creationDate.toLocaleString('default', { month: 'long' })} ${creationDate.getFullYear()}`;
    const timeSinceString = `(${timezSince2(creationDate)} ago)`
    document.getElementById("userCreationDate").textContent = `${dateString} ${timeSinceString}`;
    
    document.getElementById("userSteamID").textContent = userObj.User.accdetails.steamid;
    document.getElementById("userAvatar").src = userObj.User.accdetails.avatar.split(".jpg").join("") + "_full.jpg";

    document.getElementById("userUsername").classList.remove("skeleton");
    document.getElementById("userRealName").classList.remove("skeleton");
    document.getElementById("userFriendCount").classList.remove("skeleton");
    document.getElementById("userCreationDate").classList.remove("skeleton");
    document.getElementById("userSteamID").classList.remove("skeleton");
    document.getElementById("userAvatar").classList.remove("skeleton");
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

function fillInTopStatBoxes(graphData) {
    const UNCountries = 195;
    let uniqueCountryCodes = extractUniqueCountryCodesFromFriends(graphData.usergraphdata.frienddetails)

    document.getElementById("statBoxFriendCount").textContent = graphData.usergraphdata.userdetails.User.friendids.length;
    document.getElementById("statBoxUniqueCountries").textContent = uniqueCountryCodes.length;
    document.getElementById("statBoxGlobalCoverage").textContent = Math.floor((uniqueCountryCodes.length/195)*100) + "%";
    document.getElementById("statBoxDictatorships").textContent = ruledByDictatorCountries(uniqueCountryCodes)

    removeSkeletonClasses(["statBoxFriendCount", "statBoxUniqueCountries", 
            "statBoxGlobalCoverage", "statBoxDictatorships"])
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
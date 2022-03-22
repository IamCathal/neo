"use strict";

export function getCrawlIDs() {
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

export function timezSince(targetDate) {
    let seconds = Math.floor((new Date()-targetDate)/1000)
    let interval = seconds / 31536000 
    if (interval > 1) {
        return Math.floor(interval) + "y ago";
    }
    interval = seconds / 2592000; // months
    if (interval > 1) {
        return Math.floor(interval) + "m ago";
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
    return Math.floor(seconds) + "s ago";
}

export function monthsSince(timestamp) {
    const timeObj = new Date(timestamp * 1000)
    let monthDiff;
    const currTime = new Date();
    monthDiff = (currTime.getFullYear() - timeObj.getFullYear()) * 12
    monthDiff += currTime.getMonth()
    monthDiff -= timeObj.getMonth()
    return monthDiff <= 0 ? 0 : monthDiff
}

export function removeSkeletonClasses(elementIDs) {
    elementIDs.forEach(ID => {
        document.getElementById(ID).classList.remove("skeleton");
        document.getElementById(ID).classList.remove("skeleton-text");
    })
}

export function intToMonth(month) {
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

// https://dev.to/jorik/country-code-to-flag-emoji-a21
export function getFlagEmoji(countryCode) {
    const codePoints = countryCode
      .toUpperCase()
      .split('')
      .map(char =>  127397 + char.charCodeAt());
    return String.fromCodePoint(...codePoints);
}

export function countryCodeToName(code) {
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

export const continents = {
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
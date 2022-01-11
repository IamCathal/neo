function newUserEvents() {
    this.events = [];
}

newUserEvents.prototype.dequeue = function () { return this.events.shift() };
newUserEvents.prototype.enqueue = function (e) { return this.events.push(e) };
newUserEvents.prototype.isEmpty = function () { return this.events.length == 0 };
newUserEvents.prototype.getAll = function () { return this.events; }

  
function initAndMonitorWebsocket() {
    let currentUserEvents = new newUserEvents();
    const animationIntroClasses = "animated animatedFadeInFromLeft fadeInFromLeft"
    const animationNormalClasses = "animated animatedInFromLeft inFromLeft";
    const animationOutroClasses = "animated animatedOutFromLeft fadeOutFromLeft"

    let wsConn = new WebSocket("ws://localhost:2590/ws/newuserstream");
    wsConn.addEventListener("close", (evt) => {
        console.log("CLOSED!");
    })

    wsConn.addEventListener("message", (evt) => {
        if (currentUserEvents.events.length == 8) {
            let event = currentUserEvents.dequeue()
        }

        newUser = JSON.parse(evt.data);
        newUser.avatar = newUser.avatar.split(".jpg").join("") + "_full.jpg"
        currentUserEvents.enqueue(newUser)
     

        let allEventsContent = [];
        let i = 0
        currentUserEvents.getAll().forEach(event => {
            if (i == 0) {
                allEventsContent.unshift(marshalIntoHTML(event, animationOutroClasses))
            } else if (i == currentUserEvents.getAll().length - 1) {
                allEventsContent.unshift(marshalIntoHTML(event, animationIntroClasses))
            } else {
                allEventsContent.unshift(marshalIntoHTML(event, animationNormalClasses))
            }
            i++
        })
        let allEventsHTML = allEventsContent.join("")

        document.getElementById("newUserDiv").innerHTML = "";
        document.getElementById("newUserDiv").innerHTML += allEventsHTML;

        setInterval(() => {
            currentUserEvents.getAll().forEach(foundUser => {
                document.getElementById(foundUser.steamid).textContent = timeSince(new Date(foundUser.crawltime* 1000))
            });
        }, 1000);
    })
}

function timeSince(targetDate) {
    let seconds = Math.floor((new Date()-targetDate)/1000)
    let interval = seconds / 31536000 // years
    interval = seconds / 2592000; // months
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

function marshalIntoHTML(newUser, animationClasses) {
    // console.log(`marshaling ${newUser.personaname}`)
    return `
    <div class="col-auto newUserEventBox ${animationClasses}">
        <div class="row-2" style="height: 11rem; width: 9rem">
            <div class="row" >
            <div class="col pt-2">
                <a href="${newUser.profileurl}"> 
                    <img 
                        src="${newUser.avatar}"
                        style="height: 8.5rem; padding-left: 0.5rem"
                    >
                </a>
            </div>
            </div>
            <div class="row" style="width: 10.7rem; padding-left: 0.5rem">
                <div class="col truncate" style="font-size: 0.9rem;">
                    <span class="bolder"> ${newUser.personaname}</span>
                </div>
            </div>
            <div class="row">
                <div class="col" style="width: 8rem; font-size: 0.9rem;">
                    <p style="float: right" 
                        class="pr-1" 
                        id="${newUser.steamid}" > 
                        6s ago
                    </p>
                </div>
            </div>
        </div>
    </div>
    `
}
  
//   let q = new Queue();
//   for (let i = 1; i <= 7; i++) {
//     q.enqueue(i);
//   }
//   // get the current item at the front of the queue
//   console.log(q.peek()); // 1
  
//   // get the current length of queue
//   console.log(q.length()); // 7
  
//   // dequeue all elements
//   while (!q.isEmpty()) {
//       console.log(q.dequeue());
//   }
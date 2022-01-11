function newUserEvents() {
    this.events = [];
}

newUserEvents.prototype.dequeue = function () { return this.events.pop()};
newUserEvents.prototype.enqueue = function (e) { return this.events.unshift(e)};
newUserEvents.prototype.isEmpty = function () { return this.events.length == 0 };
newUserEvents.prototype.peek = function () { return !this.isEmpty() ? this.events[0] : undefined};
newUserEvents.prototype.getAll = function () { return this.events; }
newUserEvents.prototype.length = function () {
    return this.events.length;
};

  
function initAndMonitorWebsocket() {
    let currentUserEvents = new newUserEvents();

    let wsConn = new WebSocket("ws://localhost:2590/ws/newuserstream");
    wsConn.addEventListener("close", (evt) => {
        console.log("CLOSED!");
        document.getElementById("newUserDiv").textContent += "CONNECTION CLOSED";
    })

    wsConn.addEventListener("message", (evt) => {
        console.log(currentUserEvents.events.length)
        if (currentUserEvents.events.length == 8) {
            currentUserEvents.dequeue()
        }

        newUser = JSON.parse(evt.data);
        fullResolutionProfiler = newUser.avatar.split(".jpg").join("") + "_full.jpg"
        newUserBox = `
        <div class="col-auto newUserEventBox">
            <div class="row-2" style="height: 11rem; width: 9rem">
                <div class="row" >
                <div class="col pt-2">
                    <a href="${newUser.profileurl}"> 
                        <img 
                            src="${fullResolutionProfiler}"
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
                        <p style="float: right" class="pr-1"> 6s ago </p>
                    </div>
                </div>
            </div>
        </div>

        `
        currentUserEvents.enqueue(newUserBox)

        allEventsContent = currentUserEvents.getAll().join(" ")
        document.getElementById("newUserDiv").innerHTML = "";
        document.getElementById("newUserDiv").innerHTML += allEventsContent;
    })
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
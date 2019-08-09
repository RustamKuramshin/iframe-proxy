(function(open) {
    XMLHttpRequest.prototype.open = function(method, url, async, user, pass) {
        open.call(this, method, "/xhrproxy" + url, async, user, pass);
        this.setRequestHeader("X-Mark", "to-root");
    };
})(XMLHttpRequest.prototype.open);
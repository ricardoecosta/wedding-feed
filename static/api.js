const TIMEOUT_MS = 30000;
const LONG_TIMEOUT_MS = 60000;

var Api = function ($) {
    this.$ = $;
    messagesResourceLocation = "/api/messages";
};

Api.prototype.fetchMessages = function () {
    var deferred = this.$.Deferred();

    this.$.ajax({
        url: messagesResourceLocation,
        type: 'GET',
        crossDomain: true,
        cache: false,
        timeout: LONG_TIMEOUT_MS
    }).then(function (data) {
        if (data) {
            deferred.resolve(JSON.parse(data));
        }
        deferred.resolve([]);
    }).fail(function (xhr) {
        deferred.reject(xhr.status);
    });

    return deferred.promise();
};

Api.prototype.unarchivedMessages = function () {
    var deferred = this.$.Deferred();

    this.$.ajax({
        url: `${messagesResourceLocation}/unarchived`,
        type: 'GET',
        crossDomain: true,
        cache: false,
        timeout: LONG_TIMEOUT_MS
    }).then(function (data) {
        if (data) {
            deferred.resolve(JSON.parse(data));
        }
        deferred.resolve([]);
    }).fail(function (xhr) {
        deferred.reject(xhr.status);
    });

    return deferred.promise();
};

Api.prototype.archiveMessage = function (messageId) {
    var deferred = this.$.Deferred();

    this.$.ajax({
        url: `${messagesResourceLocation}/${messageId}/archive`,
        type: 'PUT',
        crossDomain: true,
        cache: false,
        timeout: TIMEOUT_MS
    }).then(function () {
        deferred.resolve();
    }).fail(function (xhr) {
        deferred.reject(xhr.status);
    });

    return deferred.promise();
};

Api.prototype.unarchiveMessage = function (messageId) {
    var deferred = this.$.Deferred();

    this.$.ajax({
        url: `${messagesResourceLocation}/${messageId}/unarchive`,
        type: 'PUT',
        crossDomain: true,
        cache: false,
        timeout: TIMEOUT_MS
    }).then(function () {
        deferred.resolve();
    }).fail(function (xhr) {
        deferred.reject(xhr.status);
    });

    return deferred.promise();
};

Api.prototype.createMessage = function () {
    // TODO
};
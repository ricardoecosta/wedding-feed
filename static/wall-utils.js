const WallUtils = function () {
};

WallUtils.prototype.setupFullscreen = function (button, div) {
    var fullscreenFunc = div.requestFullscreen;

    if (!fullscreenFunc) {
        ["mozRequestFullScreen", "msRequestFullscreen", "webkitRequestFullScreen"].forEach(function (req) {
            fullscreenFunc = fullscreenFunc || div[req];
        });
    }

    button.addEventListener("click", function () {
        fullscreenFunc.call(div);
    });
};

const stockPhotos = [
    {image_url: "img/stock-photo-0.jpg", image_width: 1455, image_height: 2185},
    {image_url: "img/stock-photo-1.jpg", image_width: 1098, image_height: 1742},
    {image_url: "img/stock-photo-4.jpg", image_width: 597, image_height: 770},
    {image_url: "img/stock-photo-5.jpg", image_width: 929, image_height: 1353}
];

WallUtils.prototype.stockPhotosUrls = stockPhotos.map(message => message.image_url);

WallUtils.prototype.pickRandomStockPhoto = function () {
    return stockPhotos[Math.floor(Math.random() * stockPhotos.length)];
};

WallUtils.prototype.pickRandomPhotoFilter = function () {
    return photoFilters[Math.floor(Math.random() * photoFilters.length)];
};

const photoFilters = [
    "_1977",
    "aden",
    "brannan",
    "brooklyn",
    "clarendon",
    "earlybird",
    "gingham",
    "hudson",
    "inkwell",
    "kelvin",
    "lark",
    "lofi",
    "maven",
    "mayfair",
    "moon",
    "nashville",
    "perpetua",
    "reyes",
    "rise",
    "slumber",
    "stinson",
    "toaster",
    "valencia",
    "walden",
    "willow",
    "xpro2",
];

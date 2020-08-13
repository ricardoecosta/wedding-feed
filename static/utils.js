var Utils = function () {
};

Utils.prototype.openDialog = function (title, text) {
    App.dialog({
        title: title,
        text: text,
        okButton: "Voltar"
    });
};

Utils.prototype.openYesNoDialog = function (title, text, confirmationFunction) {
    App.dialog({
        title: title,
        text: text,
        okButton: "Sim",
        cancelButton: "Não"}, function (yes) {
        if (yes) {
            confirmationFunction();
        }
    });
};

Utils.prototype.preloadImages = function(imageUrls) {
    console.info(`preloading images: ${imageUrls}`);
    imageUrls.forEach(function(imageUrl) {
        const img = new Image();
        img.src = imageUrl;
    });
};
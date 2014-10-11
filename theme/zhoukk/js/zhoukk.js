$(document).ready(function() {
    $backToTopEle = $('<div class="backToTop"></div>').appendTo($("body")).attr("title", "返回顶部").click(function() {
        $("html, body").animate({
            scrollTop: 0
        }, 120);
    }),
    $backToTopFun = function() {
        var st = $(document).scrollTop(),
            winh = $(window).height();
        (st > 0) ? $backToTopEle.show() : $backToTopEle.hide();
        //IE6下的定位
        if (!window.XMLHttpRequest) {
            $backToTopEle.css("top", st + winh - 166);
        }
    };
    $(window).bind("scroll", $backToTopFun);
    $backToTopFun();

    $(document).pjax('a', '#pjax-container')

    $(document).on('pjax:send', function() {
        $('#loading').show()
    })

    $(document).on('pjax:complete', function() {
        $('#loading').hide()
    })

    $(document).on('pjax:timeout', function(event) {
        event.preventDefault()
    })
});
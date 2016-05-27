FROM busybox

ADD slack-jobs /usr/bin/slack-jobs

EXPOSE 8765
CMD ["/usr/bin/slack-jobs"]

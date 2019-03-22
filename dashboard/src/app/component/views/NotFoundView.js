import React from "react";

import AppLayout from "/lib/component/base/AppLayout";

import NotFound from "/app/component/partial/NotFound";

class NotFoundView extends React.PureComponent {
  render() {
    return <AppLayout content={<NotFound />} />;
  }
}

export default NotFoundView;

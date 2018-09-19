import React from "react";

import AppLayout from "/components/AppLayout";
import NotFound from "/components/partials/NotFound";

class NotFoundView extends React.PureComponent {
  render() {
    return <AppLayout content={<NotFound />} />;
  }
}

export default NotFoundView;

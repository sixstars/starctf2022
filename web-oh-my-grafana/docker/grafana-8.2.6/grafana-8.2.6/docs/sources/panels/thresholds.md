+++
title = "Thresholds"
weight = 440
+++

# Thresholds

Thresholds set the color of either the value text or the background depending on conditions that you define.

You can define thresholds one of two ways:

- **Absolute** thresholds are defined based on a number. For example, 80 on a scale of 1 to 150.
- **Percentage** thresholds are defined relative to minimum or maximum. For example, 80 percent.

You can apply thresholds to most, but not all, visualizations.

## Default thresholds

On visualizations that support it, Grafana sets default threshold values of:

- 80 = red
- Base = green
- Mode = Absolute

The **Base** value represents minus infinity. It is generally the “good” color.

## Add a threshold

You can add as many thresholds to a panel as you want. Grafana automatically sorts thresholds from highest value to lowest.

1. Navigate to the panel you want to add a threshold to.
1. Open the category **Thresholds** in the right side options pane.
1. Click **Add threshold**.
1. Grafana adds a threshold with suggested numerical and color values.
1. Accept the recommendations or edit the new threshold.
   - **Edit color:** Click the color dot you wish to change and then select a new color.
   - **Edit number:** Click the number you wish to change and then enter a new number.
   - **Thresholds mode -** Click the mode to change it for all thresholds on this panel.
1. Click **Save** to save the changes in the dashboard.

## Add a threshold to a Graph (old) panel

In the Graph panel visualization, thresholds allow you to add arbitrary lines or sections to the graph to make it easier to see when the graph crosses a particular threshold.

1. Navigate to the graph panel you want to add a threshold to.
1. On the Panel tab, click **Thresholds**.
1. Click **Add threshold**.
1. Fill in as many fields as you want. Only the **T1** fields are required.
   - **T1 -** Both values are required to display a threshold.
     - **lt** or **gt** - Select **lt** for less than or **gt** for greater than to indicate what the threshold applies to.
     - **Value -** Enter a threshold value. Grafana draws a threshold line along the Y-axis at that value.
   - **Color -** Choose a condition that corresponds to a color, or define your own color.
     - **custom -** You define the fill color and line color.
     - **critical -** Fill and line color are red.
     - **warning -** Fill and line color are yellow.
     - **ok -** Fill and line color are green.
   - **Fill -** Controls whether the threshold fill is displayed.
   - **Line -** Controls whether the threshold line is displayed.
   - **Y-Axis -** Choose **left** or **right**.
1. Click **Save** to save the changes in the dashboard.

## Delete a threshold

1. Navigate to the panel you want to add a threshold to.
1. Click the **Field** tab. (Or **Panel** tab for a graph panel.)
1. Click the trash can icon next to the threshold you want to remove.
1. Click **Save** to save the changes in the dashboard.

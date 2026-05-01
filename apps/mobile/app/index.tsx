import "@/global.css"

import { View, StyleSheet, SafeAreaView, StatusBar } from "react-native";
import { useState } from "react";
import DashboardHeader from "./components/DashboardHeader";
import SearchBar from "./components/SearchBar";
import StartWorkoutCard from "./components/StartWorkoutCard";
import RecentActivity from "./components/RecentActivity";
import BottomMenu from "./components/BottomMenu";

type TabName = 'home' | 'explore' | 'stats' | 'profile';

export default function App() {
  const [activeTab, setActiveTab] = useState<TabName>('home');

  return (
    <SafeAreaView style={styles.safe}>
      <StatusBar barStyle="light-content" backgroundColor="#111113" />
      <View style={styles.container}>
        <DashboardHeader />
        <SearchBar />
        <StartWorkoutCard />
        <RecentActivity />
      </View>
      <BottomMenu activeTab={activeTab} onTabPress={setActiveTab} />
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: {
    flex: 1,
    backgroundColor: '#111113',
  },
  container: {
    flex: 1,
    backgroundColor: '#111113',
  },
});
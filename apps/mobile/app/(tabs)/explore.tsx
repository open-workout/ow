import { View, Text, ScrollView, TouchableOpacity, TextInput, StatusBar } from "react-native";
import { SafeAreaView } from 'react-native-safe-area-context';
import { Image } from 'expo-image';
import { Ionicons, MaterialCommunityIcons } from '@expo/vector-icons';
import { useState } from 'react';

const categories = ['All', 'Chest', 'Back', 'Legs', 'Arms', 'Shoulders'];

const recentSearches = [
  {
    name: 'Barbell Bench Press',
    muscle: 'Chest',
    equipment: 'Barbell',
    image: 'https://images.unsplash.com/photo-1581009146145-b5ef050c2e1e?auto=format&fit=crop&w=150&h=150&q=80',
  },
  {
    name: 'Incline Dumbbell Press',
    muscle: 'Chest',
    equipment: 'Dumbbell',
    image: 'https://images.unsplash.com/photo-1541534741688-6078c6bfb5c5?auto=format&fit=crop&w=150&h=150&q=80',
  },
];

const popular = [
  { name: 'Squat (Barbell)', muscle: 'Legs', equipment: 'Barbell' },
  { name: 'Pull Up', muscle: 'Back', equipment: 'Bodyweight' },
  {
    name: 'Deadlift (Barbell)',
    muscle: 'Back / Legs',
    equipment: 'Barbell',
    image: 'https://images.unsplash.com/photo-1599058917212-d750089bc07e?auto=format&fit=crop&w=150&h=150&q=80',
  },
];

export default function ExploreScreen() {
  const [activeCategory, setActiveCategory] = useState(0);

  return (
    <SafeAreaView style={{ flex: 1, backgroundColor: '#0a0a0a' }} edges={['top']}>
      <StatusBar barStyle="light-content" backgroundColor="#0a0a0a" />

      {/* Header */}
      <View style={{ paddingHorizontal: 24, paddingTop: 8, paddingBottom: 0 }}>
        <View style={{ flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', marginBottom: 16 }}>
          <Text style={{ color: '#fff', fontSize: 24, fontWeight: '700', letterSpacing: -0.3 }}>Library</Text>
          <TouchableOpacity style={{ width: 40, height: 40, borderRadius: 20, backgroundColor: '#18181b', alignItems: 'center', justifyContent: 'center', borderWidth: 1, borderColor: '#27272a' }}>
            <Ionicons name="options-outline" size={20} color="#a1a1aa" />
          </TouchableOpacity>
        </View>

        {/* Search */}
        <View style={{ flexDirection: 'row', alignItems: 'center', backgroundColor: '#18181b', borderWidth: 1, borderColor: '#27272a', borderRadius: 12, paddingHorizontal: 14, paddingVertical: 12, marginBottom: 16 }}>
          <Ionicons name="search-outline" size={18} color="#52525b" style={{ marginRight: 10 }} />
          <TextInput
            placeholder="Find an exercise..."
            placeholderTextColor="#52525b"
            style={{ flex: 1, color: '#fff', fontSize: 14 }}
          />
        </View>

        {/* Category pills */}
        <ScrollView
          horizontal
          showsHorizontalScrollIndicator={false}
          contentContainerStyle={{ paddingBottom: 12, gap: 8 }}
          style={{ borderBottomWidth: 1, borderBottomColor: '#18181b' }}
        >
          {categories.map((cat, i) => (
            <TouchableOpacity
              key={cat}
              onPress={() => setActiveCategory(i)}
              style={{
                paddingHorizontal: 20,
                paddingVertical: 10,
                borderRadius: 999,
                backgroundColor: i === activeCategory ? '#f4f4f5' : '#18181b',
                borderWidth: i === activeCategory ? 0 : 1,
                borderColor: '#27272a',
              }}
            >
              <Text style={{
                fontSize: 14,
                fontWeight: '600',
                color: i === activeCategory ? '#09090b' : '#d4d4d8',
              }}>
                {cat}
              </Text>
            </TouchableOpacity>
          ))}
        </ScrollView>
      </View>

      {/* List */}
      <ScrollView
        style={{ flex: 1 }}
        showsVerticalScrollIndicator={false}
        contentContainerStyle={{ paddingHorizontal: 24, paddingBottom: 24 }}
      >
        <Text style={{ color: '#71717a', fontSize: 11, fontWeight: '700', textTransform: 'uppercase', letterSpacing: 1, marginTop: 16, marginBottom: 12 }}>
          Recently Searched
        </Text>
        {recentSearches.map((ex) => <ExerciseRow key={ex.name} {...ex} />)}

        <Text style={{ color: '#71717a', fontSize: 11, fontWeight: '700', textTransform: 'uppercase', letterSpacing: 1, marginTop: 24, marginBottom: 12 }}>
          Popular
        </Text>
        {popular.map((ex) => <ExerciseRow key={ex.name} {...ex} />)}
      </ScrollView>
    </SafeAreaView>
  );
}

function ExerciseRow({ name, muscle, equipment, image }: { name: string; muscle: string; equipment: string; image?: string }) {
  return (
    <View style={{
      flexDirection: 'row',
      alignItems: 'center',
      backgroundColor: 'rgba(24,24,27,0.4)',
      borderWidth: 1,
      borderColor: 'rgba(39,39,42,0.5)',
      borderRadius: 16,
      padding: 12,
      marginBottom: 12,
    }}>
      {image ? (
        <Image
          source={{ uri: image }}
          style={{ width: 64, height: 64, borderRadius: 12, borderWidth: 1, borderColor: '#27272a' }}
          contentFit="cover"
        />
      ) : (
        <View style={{ width: 64, height: 64, borderRadius: 12, backgroundColor: '#27272a', borderWidth: 1, borderColor: '#3f3f46', alignItems: 'center', justifyContent: 'center' }}>
          <MaterialCommunityIcons name="dumbbell" size={24} color="#a1a1aa" />
        </View>
      )}
      <View style={{ flex: 1, marginLeft: 16 }}>
        <Text style={{ color: '#f4f4f5', fontSize: 14, fontWeight: '700' }}>{name}</Text>
        <Text style={{ color: '#52525b', fontSize: 12, marginTop: 4 }}>{muscle} • {equipment}</Text>
      </View>
      <TouchableOpacity style={{ width: 32, height: 32, borderRadius: 16, backgroundColor: '#27272a', alignItems: 'center', justifyContent: 'center' }}>
        <Ionicons name="add" size={18} color="#d4d4d8" />
      </TouchableOpacity>
    </View>
  );
}
